import requests
from bs4 import BeautifulSoup
from typing import Optional
import database
from models import IndiaCodeAct
import os
from multiprocessing import Pool

BASE_URL = "https://www.indiacode.nic.in"
ACTS_URL = "https://www.indiacode.nic.in/handle/123456789/1362/browse?nccharset=4A8500CD&type=shorttitle&sort_by=3&order=ASC&rpp=877&submit_browse=Update"

class IndiaCodeActsScraper:
    """A scraper for extracting links from the India Code website."""

    def __init__(
        self,
        base_url: str = "https://www.indiacode.nic.in/handle/123456789/1362/browse?nccharset=4A8500CD&type=shorttitle&sort_by=3&order=ASC&rpp=877&submit_browse=Update",
    ):
        self.base_url = base_url
        self.headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
        }
        self.session = requests.Session()
        self.session.headers.update(self.headers)

    def fetch_page(self, url: str) -> Optional[BeautifulSoup]:
        """Fetch and parse a webpage."""
        try:
            print(f"Fetching: {url}")
            response = self.session.get(url, timeout=60)
            response.raise_for_status()
            return BeautifulSoup(response.content, "html.parser")
        except requests.exceptions.RequestException as e:
            print(f"Error fetching {url}: {e}")
            return None

    def extract_links(self, url: str) -> list[IndiaCodeAct]:
        """
        Extract all 'view...' links from the India Code website

        Args:
            url (str): The URL of the India Code browse page

        Returns:
            List[Dict[str, str]]: A list of dictionaries containing act titles and their view links
        """
        soup = self.fetch_page(url)
        if not soup:
            return []

        # Find the table containing the data
        table = soup.find("table", {"class": "table"})
        if not table:
            print("Could not find data table on the page")
            return []

        # Extract table rows
        rows = table.find_all("tr")
        if len(rows) <= 1:
            print("No data rows found in the table")
            return []

        # Extract data from each row
        acts_data: list[IndiaCodeAct] = []
        for i, row in enumerate(rows[:]):  # Skip header row
            cols = row.find_all("td")
            if len(cols) >= 3:
                # Get the enactment date from the first column
                enactment_date = cols[0].get_text(strip=True)

                # Get the act number from the second column
                act_number = cols[1].get_text(strip=True)

                # Short Title of the Act
                short_title = cols[2].get_text(strip=True)

                # Get the view link from the third column
                view_link_tag = cols[3].find("a")
                if view_link_tag and "href" in view_link_tag.attrs:
                    view_link = view_link_tag["href"]
                    # Make sure the link is absolute
                    if view_link.startswith("/"):
                        view_link = BASE_URL + view_link

                    acts_data.append(
                        IndiaCodeAct(
                            act_number=act_number,
                            title=short_title,
                            view_link=view_link,
                            enactment_date=enactment_date,
                        )
                    )

        print(f"Successfully extracted {len(acts_data)} links")
        return acts_data

    def fetch_pdf_link(self, act_link: str) -> Optional[str]:
        """Fetch the PDF link from the act page using the CSS selector."""
        try:
            response = self.session.get(act_link, timeout=60)
            response.raise_for_status()
            soup = BeautifulSoup(response.content, "html.parser")
            a_tag = soup.select_one("div.row > a:first-child")
            if a_tag and a_tag.get("href", "").endswith(".pdf"):
                pdf_url = a_tag["href"]
                if pdf_url.startswith("/"):
                    pdf_url = BASE_URL + pdf_url
                return pdf_url
        except Exception as e:
            print(f"Error fetching PDF link from {act_link}: {e}")
        return None

    def save_pdf_link_to_db(self, db_path: str, act_link: str):
        """Fetch PDF link and save to DB if found."""
        db = database.SQLiteManager(db_path)
        pdf_url = self.fetch_pdf_link(act_link)
        if pdf_url:
            db.insert_pdf_link(act_link, pdf_url)

    def prepare_pdf_links(self, db_path: str, processes: int = 8):
        """Scan all act links and save their PDF links to DB using multiprocessing."""
        db = database.SQLiteManager(db_path)
        links = db.get_all_links()
        print(f"Total act links to process: {len(links)}")
        args = [(db_path, link) for link in links]
        with Pool(processes=processes) as pool:
            pool.starmap(self.save_pdf_link_to_db, args)

    def download_pdf(self, args):
        """Download a single PDF file given the act link and pdf url."""
        act_link, pdf_url, output_dir, db_path = args
        db = database.SQLiteManager(db_path)
        filename = pdf_url.split("/")[-1]
        filepath = os.path.join(output_dir, filename)
        try:
            r = self.session.get(pdf_url, timeout=60)
            r.raise_for_status()
            with open(filepath, "wb") as f:
                f.write(r.content)
            print(f"Saved PDF: {filepath}")
            db.update_pdf_status(pdf_url, "success")
        except Exception as e:
            print(f"Error downloading PDF {pdf_url}: {e}")
            db.update_pdf_status(pdf_url, "failed", str(e))

    def download_all_pdfs(self, db_path: str, output_dir: str = "pdfs", processes: int = 8):
        """Download PDFs in parallel for pending/failed ones only."""
        db = database.SQLiteManager(db_path)
        pdf_entries = db.get_pending_or_failed_pdfs()
        if not os.path.exists(output_dir):
            os.makedirs(output_dir)
        args = [(entry["act_view_link"], entry["pdf_url"], output_dir, db_path) for entry in pdf_entries]
        if args:
            with Pool(processes=processes) as pool:
                pool.map(self.download_pdf, args)
        else:
            print("No pending or failed PDFs to download.")


def scrape_all_acts(url: str) -> list[IndiaCodeAct]:
    """Scrape all acts from the given URL."""
    scraper = IndiaCodeActsScraper()
    # Get the links
    acts_data = scraper.extract_links(url)

    print(f"Total acts extracted: {len(acts_data)}")

    if acts_data:
        # Display first few results
        print("\nFirst 5 extracted links:")
        for i, act in enumerate(acts_data[:5]):
            print(f"{act.act_number}, {act.title}")
            print(f"   Link: {act.view_link}\n")

        # Extract links from saved HTML and store in DB (example paths, update as needed)
        db = database.SQLiteManager("act_links.db")
        db.save_acts(acts_data)
    else:
        print("No data was extracted")


def save_all_pdf_links():
    # Create scraper instance
    scraper = IndiaCodeActsScraper()
    # Run PDF link preparation in parallel
    scraper.prepare_pdf_links("act_links.db", processes=8)

def download_all_pdfs():
    scraper = IndiaCodeActsScraper()
    scraper.download_all_pdfs("act_links.db", "pdfs", processes=8)


def main():
    """Main function to run the scraper."""
    # scrape_all_acts(ACTS_URL)
    # save_all_pdf_links()

    download_all_pdfs()
    


if __name__ == "__main__":
    main()
