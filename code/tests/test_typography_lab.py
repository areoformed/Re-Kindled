#!/usr/bin/env python3
from __future__ import annotations

import json
import unittest
from html.parser import HTMLParser
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


class InventoryParser(HTMLParser):
    def __init__(self) -> None:
        super().__init__()
        self.ids: set[str] = set()
        self.labels: set[str] = set()

    def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
        values = dict(attrs)
        if values.get("id"):
            self.ids.add(values["id"] or "")
        if tag == "label" and values.get("for"):
            self.labels.add(values["for"] or "")


class TypographyLabTest(unittest.TestCase):
    def test_lab_has_accessible_dials_and_export(self) -> None:
        source = (ROOT / "typography-lab" / "index.html").read_text(encoding="utf-8")
        parser = InventoryParser()
        parser.feed(source)
        dials = {"bodyPixels", "baselinePitch", "distance", "sideMargin"}
        self.assertTrue(dials.issubset(parser.ids))
        self.assertTrue(dials.issubset(parser.labels))
        self.assertIn("exportButton", parser.ids)
        self.assertIn("minimumLineGapUnits", source)
        self.assertIn("applyQueryDials", source)
        self.assertIn("prefers-reduced-motion", source)

    def test_reference_recipe_matches_physical_trial(self) -> None:
        recipe = json.loads((ROOT / "typography-lab" / "reference-recipe.json").read_text(encoding="utf-8"))
        self.assertEqual(recipe["candidate"]["body_fbink_pixels"], 49)
        self.assertEqual(recipe["candidate"]["baseline_pitch_pixels"], 57)
        self.assertEqual(recipe["candidate"]["line_gap_pixels"], 8)


if __name__ == "__main__":
    unittest.main()
