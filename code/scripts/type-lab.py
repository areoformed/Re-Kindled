#!/usr/bin/env python3
"""Calculate ReKindled type geometry without modifying a font or device."""

from __future__ import annotations

import argparse
import json
import math
from pathlib import Path
from typing import Any


def unspaced_height(face: dict[str, Any], pixels: int) -> int:
    vertical = face["ascent_units"] - face["descent_units"]
    baseline = math.ceil(pixels * face["ascent_units"] / vertical)
    descent = math.ceil(pixels * face["descent_units"] / vertical)
    return baseline + abs(descent)


def line_geometry(face: dict[str, Any], pixels: int) -> tuple[int, int]:
    vertical = face["ascent_units"] - face["descent_units"]
    gap = math.ceil(pixels * face.get("line_gap_units", 0) / vertical)
    return unspaced_height(face, pixels) + gap, gap


def minimum_line_gap_units(gap_pixels: int, pixels: int, face: dict[str, Any]) -> int:
    if gap_pixels <= 0:
        return 0
    vertical = face["ascent_units"] - face["descent_units"]
    return math.floor((gap_pixels - 1) * vertical / pixels) + 1


def pixels_for_angle(angle: float, distance: float, ppi: float, face: dict[str, Any]) -> int:
    cap_pixels = 2 * distance * ppi * math.tan(math.radians(angle) / 2)
    vertical = face["ascent_units"] - face["descent_units"]
    return math.ceil(cap_pixels * vertical / face["cap_height_units"])


def achieved_angle(pixels: int, distance: float, ppi: float, face: dict[str, Any]) -> float:
    vertical = face["ascent_units"] - face["descent_units"]
    cap_inches = (pixels * face["cap_height_units"] / vertical) / ppi
    return math.degrees(2 * math.atan(cap_inches / (2 * distance)))


def calculate(
    preset: dict[str, Any],
    *,
    body_change: float,
    target_angle: float | None,
    pitch_change: float,
    distance: float | None,
) -> dict[str, Any]:
    face = preset["fonts"]["narrative"]
    calibration = preset["calibration"]
    ppi = float(calibration["panel_pixels_per_inch"])
    distance = float(distance or calibration["viewing_distance_inches"])
    current_pixels = int(face["fbink_pixels"])
    current_pitch, current_gap = line_geometry(face, current_pixels)

    candidate_pixels = round(current_pixels * (1 + body_change / 100))
    if target_angle is not None:
        candidate_pixels = pixels_for_angle(target_angle, distance, ppi, face)
    if candidate_pixels < 8:
        raise ValueError("candidate body size is below 8 pixels")

    target_pitch = round(current_pitch * (1 + pitch_change / 100))
    base_height = unspaced_height(face, candidate_pixels)
    requested_gap = max(0, target_pitch - base_height)
    gap_units = minimum_line_gap_units(requested_gap, candidate_pixels, face)
    candidate_face = dict(face, line_gap_units=gap_units)
    candidate_pitch, candidate_gap = line_geometry(candidate_face, candidate_pixels)
    angle = achieved_angle(candidate_pixels, distance, ppi, face)

    return {
        "schema_version": 1,
        "source_preset": preset.get("id", "unknown"),
        "current": {
            "body_fbink_pixels": current_pixels,
            "line_gap_pixels": current_gap,
            "baseline_pitch_pixels": current_pitch,
        },
        "candidate": {
            "body_fbink_pixels": candidate_pixels,
            "line_gap_pixels": candidate_gap,
            "baseline_pitch_pixels": candidate_pitch,
            "achieved_cap_height_degrees": angle,
            "viewing_distance_inches": distance,
            "above_nasa_025_degree_minimum": angle >= 0.25,
        },
        "actual_change_percent": {
            "body_size": 100 * (candidate_pixels / current_pixels - 1),
            "line_pitch": 100 * (candidate_pitch / current_pitch - 1),
        },
        "recipe": {
            "narrative_fbink_pixels": candidate_pixels,
            "hhea_line_gap_units": gap_units,
            "source_preset": preset.get("id", "unknown"),
        },
    }


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Turn physical typography dials into an FBInk/OpenType recipe."
    )
    parser.add_argument("preset", type=Path, help="source typography preset JSON")
    group = parser.add_mutually_exclusive_group()
    group.add_argument(
        "--body-change-percent",
        type=float,
        default=0,
        help="relative change to FBInk body pixels; -5 means five percent smaller",
    )
    group.add_argument(
        "--target-cap-height-degrees",
        type=float,
        help="choose body pixels from an angular cap-height target",
    )
    parser.add_argument(
        "--line-pitch-change-percent",
        type=float,
        default=0,
        help="change absolute baseline pitch relative to the source preset",
    )
    parser.add_argument("--viewing-distance-inches", type=float)
    parser.add_argument("--write-recipe", type=Path, help="also write the result as JSON")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    if not -40 <= args.body_change_percent <= 100:
        raise SystemExit("--body-change-percent must be between -40 and 100")
    if not -20 <= args.line_pitch_change_percent <= 100:
        raise SystemExit("--line-pitch-change-percent must be between -20 and 100")
    if args.target_cap_height_degrees is not None and not 0.15 <= args.target_cap_height_degrees <= 1.5:
        raise SystemExit("--target-cap-height-degrees must be between 0.15 and 1.5")
    if args.viewing_distance_inches is not None and not 6 <= args.viewing_distance_inches <= 60:
        raise SystemExit("--viewing-distance-inches must be between 6 and 60")

    preset = json.loads(args.preset.read_text(encoding="utf-8"))
    try:
        result = calculate(
            preset,
            body_change=args.body_change_percent,
            target_angle=args.target_cap_height_degrees,
            pitch_change=args.line_pitch_change_percent,
            distance=args.viewing_distance_inches,
        )
    except (KeyError, TypeError, ValueError, ZeroDivisionError) as exc:
        raise SystemExit(f"invalid preset or dial value: {exc}") from exc

    encoded = json.dumps(result, indent=2, sort_keys=True) + "\n"
    print(encoded, end="")
    if args.write_recipe:
        args.write_recipe.parent.mkdir(parents=True, exist_ok=True)
        args.write_recipe.write_text(encoded, encoding="utf-8")


if __name__ == "__main__":
    main()
