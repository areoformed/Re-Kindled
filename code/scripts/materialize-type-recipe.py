#!/usr/bin/env python3
"""Apply a ReKindled type recipe to a local font and preset."""

from __future__ import annotations

import argparse
import json
import math
import re
from pathlib import Path
from typing import Any


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Create a locally renamed font plus Kindle preset from a type-lab recipe."
    )
    parser.add_argument("--source-font", type=Path, required=True)
    parser.add_argument("--source-preset", type=Path, required=True)
    parser.add_argument("--recipe", type=Path, required=True)
    parser.add_argument("--output-font", type=Path, required=True)
    parser.add_argument("--output-preset", type=Path, required=True)
    parser.add_argument("--preset-id", default="rekindled-mono-air")
    parser.add_argument("--label", default="ReKindled Mono Air")
    parser.add_argument("--header", default="ReKindled / Mono Air")
    parser.add_argument(
        "--derived-family",
        default="ReKindled Mono Air",
        help="new internal family name; required by licenses that reserve the source name",
    )
    return parser.parse_args()


def set_name(name_table: Any, name_id: int, value: str) -> None:
    name_table.setName(value, name_id, 3, 1, 0x0409)
    try:
        name_table.setName(value, name_id, 1, 0, 0)
    except UnicodeEncodeError:
        pass


def rename_font(font: Any, family: str) -> str:
    postscript = re.sub(r"[^A-Za-z0-9-]", "", family.replace(" ", "")) + "-Regular"
    table = font["name"]
    replaced_ids = {1, 2, 3, 4, 6, 8, 16, 17, 18}
    table.names = [record for record in table.names if record.nameID not in replaced_ids]
    set_name(table, 1, family)
    set_name(table, 2, "Regular")
    set_name(table, 3, f"{family}; Regular; ReKindled local derivative")
    set_name(table, 4, f"{family} Regular")
    set_name(table, 6, postscript)
    set_name(table, 8, "ReKindled local derivative")
    set_name(table, 16, family)
    set_name(table, 17, "Regular")
    set_name(table, 18, f"{family} Regular")
    return postscript


def achieved_angle(face: dict[str, Any], distance: float, ppi: float) -> float:
    vertical = face["ascent_units"] - face["descent_units"]
    cap_inches = (face["fbink_pixels"] * face["cap_height_units"] / vertical) / ppi
    return math.degrees(2 * math.atan(cap_inches / (2 * distance)))


def main() -> None:
    try:
        from fontTools.ttLib import TTFont
    except ImportError as exc:
        raise SystemExit(
            "fontTools is required: python3 -m pip install -r requirements-fonts.txt"
        ) from exc

    args = parse_args()
    if not args.derived_family.strip():
        raise SystemExit("--derived-family cannot be empty")

    source_preset = json.loads(args.source_preset.read_text(encoding="utf-8"))
    result = json.loads(args.recipe.read_text(encoding="utf-8"))
    recipe = result.get("recipe", result)
    try:
        pixels = int(recipe["narrative_fbink_pixels"])
        line_gap_units = int(recipe["hhea_line_gap_units"])
    except (KeyError, TypeError, ValueError) as exc:
        raise SystemExit(f"recipe lacks integer body pixels or hhea line gap: {exc}") from exc
    if pixels <= 0 or line_gap_units < 0:
        raise SystemExit("recipe pixels must be positive and line gap must be non-negative")

    font = TTFont(args.source_font, recalcBBoxes=True, recalcTimestamp=True)
    if "hhea" not in font or "name" not in font:
        raise SystemExit("source font must contain hhea and name tables")
    font["hhea"].lineGap = line_gap_units
    if "DSIG" in font:
        del font["DSIG"]  # The source signature cannot remain valid after modification.
    postscript_name = rename_font(font, args.derived_family.strip())
    args.output_font.parent.mkdir(parents=True, exist_ok=True)
    font.save(args.output_font, reorderTables=True)

    preset = source_preset
    preset["id"] = args.preset_id
    preset["label"] = args.label
    preset["header"] = args.header
    narrative = preset["fonts"]["narrative"]
    narrative["path"] = f"/mnt/us/rekindled/fonts/{args.output_font.name}"
    narrative["fbink_pixels"] = pixels
    narrative["line_gap_units"] = line_gap_units
    # Use the renamed derivative for one-line utility text too; line gap has no visible effect there.
    utility = preset["fonts"].get("utility")
    if utility:
        utility["path"] = narrative["path"]

    calibration = preset["calibration"]
    distance = float(result.get("candidate", {}).get(
        "viewing_distance_inches", calibration["viewing_distance_inches"]
    ))
    ppi = float(calibration["panel_pixels_per_inch"])
    angle = achieved_angle(narrative, distance, ppi)
    candidate = result.get("candidate", {})
    gap_pixels = int(candidate.get("line_gap_pixels", 0))
    pitch_pixels = int(candidate.get("baseline_pitch_pixels", pixels + gap_pixels))
    base_height = max(1, pitch_pixels - gap_pixels)
    calibration["viewing_distance_inches"] = distance
    calibration["body_cap_height_degrees"] = angle
    metadata = preset.setdefault("local_font_metadata", {})
    metadata.update({
        "category": metadata.get("category", "unknown"),
        "license_policy": "local-source-license-controls",
        "body_fbink_pixels": pixels,
        "body_achieved_angle_degrees": angle,
        "body_line_gap_pixels": gap_pixels,
        "body_line_spacing_multiplier": pitch_pixels / base_height,
        "note": "Locally materialized derivative; verify the source license before use or redistribution.",
    })
    args.output_preset.parent.mkdir(parents=True, exist_ok=True)
    args.output_preset.write_text(json.dumps(preset, indent=2) + "\n", encoding="utf-8")

    fs_type = getattr(font.get("OS/2"), "fsType", None)
    print(json.dumps({
        "ok": True,
        "output_font": str(args.output_font),
        "output_preset": str(args.output_preset),
        "internal_postscript_name": postscript_name,
        "hhea_line_gap_units": line_gap_units,
        "source_embedding_fsType": fs_type,
        "license_notice": "You are responsible for the source font license; do not redistribute unless permitted.",
    }, indent=2))


if __name__ == "__main__":
    main()
