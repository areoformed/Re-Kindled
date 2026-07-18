#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
import sys
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


class TypeLabTest(unittest.TestCase):
    def run_lab(self, *arguments: str) -> dict:
        completed = subprocess.run(
            [
                sys.executable,
                str(ROOT / "scripts" / "type-lab.py"),
                str(ROOT / "kindle" / "reader" / "presets" / "mac" / "rekindled-mono-air.json"),
                *arguments,
            ],
            check=True,
            capture_output=True,
            text=True,
        )
        return json.loads(completed.stdout)

    def test_reference_rounding(self) -> None:
        result = self.run_lab()
        self.assertEqual(result["candidate"]["body_fbink_pixels"], 49)
        self.assertEqual(result["candidate"]["line_gap_pixels"], 8)
        self.assertEqual(result["candidate"]["baseline_pitch_pixels"], 57)
        self.assertEqual(result["recipe"]["hhea_line_gap_units"], 161)
        self.assertAlmostEqual(result["candidate"]["achieved_cap_height_degrees"], 0.324940471, places=8)

    def test_body_and_pitch_dials_remain_independent(self) -> None:
        result = self.run_lab("--body-change-percent", "-5", "--line-pitch-change-percent", "10")
        self.assertEqual(result["candidate"]["body_fbink_pixels"], 47)
        self.assertEqual(result["candidate"]["baseline_pitch_pixels"], 63)
        self.assertGreater(result["candidate"]["line_gap_pixels"], 8)
        self.assertAlmostEqual(result["actual_change_percent"]["line_pitch"], 10.526315789, places=8)

    def test_angular_target(self) -> None:
        result = self.run_lab("--target-cap-height-degrees", "0.4")
        self.assertGreaterEqual(result["candidate"]["achieved_cap_height_degrees"], 0.4)


if __name__ == "__main__":
    unittest.main()
