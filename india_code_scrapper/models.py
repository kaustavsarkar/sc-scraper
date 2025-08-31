from dataclasses import dataclass
from datetime import date

@dataclass
class IndiaCodeAct:
    act_number: int
    title: str
    view_link: str
    enactment_date: date