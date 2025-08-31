import sqlite3
from contextlib import contextmanager
from typing import List, Dict, Any, Iterator
import logging
from models import IndiaCodeAct

logger = logging.getLogger(__name__)


class SQLiteManager:
    """Manage SQLite database operations."""

    def __init__(self, db_path: str):
        self.db_path = db_path
        self._init_database()

    def _init_database(self):
        """Initialize the database with required tables."""
        with self.get_connection() as conn:
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS acts (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    act_number INTEGER NOT NULL,
                    title TEXT NOT NULL,
                    view_link TEXT NOT NULL,
                    enactment_date TEXT
                )
                """
            )
            conn.execute(
                """
                CREATE TABLE IF NOT EXISTS pdf_links (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    act_view_link TEXT NOT NULL,
                    pdf_url TEXT NOT NULL,
                    status TEXT NOT NULL DEFAULT 'pending', -- pending, success, failed
                    last_error TEXT
                )
                """
            )

    @contextmanager
    def get_connection(self) -> Iterator[sqlite3.Connection]:
        """Get a database connection with context management."""
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row  # Enable dictionary-like access
        try:
            yield conn
            conn.commit()
        except Exception as e:
            conn.rollback()
            logger.error(f"Database error: {e}")
            raise
        finally:
            conn.close()

    def save_acts(self, acts_data: List[IndiaCodeAct]) -> int:
        """Save acts data to the database."""
        if not acts_data:
            return 0

        saved_count = 0
        with self.get_connection() as conn:
            for act in acts_data:
                try:
                    conn.execute(
                        """
                        INSERT OR REPLACE INTO acts (act_number, title, view_link, enactment_date)
                        VALUES (?, ?, ?, ?)
                    """,
                        (act.act_number, act.title, act.view_link, str(act.enactment_date)),
                    )
                    saved_count += 1
                except sqlite3.Error as e:
                    logger.warning(f"Failed to save act {act.title}: {e}")

        return saved_count

    def get_all_links(self) -> list[str]:
        """Retrieve all view links from the database."""
        with self.get_connection() as conn:
            cursor = conn.execute("SELECT view_link FROM acts")
            return [row["view_link"] for row in cursor.fetchall()]

    def insert_pdf_link(self, act_view_link: str, pdf_url: str):
        """Insert a PDF link with pending status."""
        with self.get_connection() as conn:
            conn.execute(
                """
                INSERT OR IGNORE INTO pdf_links (act_view_link, pdf_url, status)
                VALUES (?, ?, 'pending')
                """,
                (act_view_link, pdf_url)
            )

    def update_pdf_status(self, pdf_url: str, status: str, error: str = None):
        """Update the download status for a PDF link."""
        with self.get_connection() as conn:
            conn.execute(
                """
                UPDATE pdf_links SET status = ?, last_error = ?
                WHERE pdf_url = ?
                """,
                (status, error, pdf_url)
            )

    def get_pending_or_failed_pdfs(self) -> List[dict]:
        """Get all PDF links with status pending or failed."""
        with self.get_connection() as conn:
            cursor = conn.execute(
                "SELECT act_view_link, pdf_url FROM pdf_links WHERE status IN ('pending', 'failed')"
            )
            return [dict(row) for row in cursor.fetchall()]