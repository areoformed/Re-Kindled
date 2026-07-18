# Physical typography and the dials

The calibration model separates values that browser CSS often collapses into one `line-height` property:

1. **Body pixels** — the `px` value sent to FBInk.
2. **Cap-height angle** — physical uppercase height as seen from the eye.
3. **Baseline pitch** — absolute panel pixels from one baseline to the next.
4. **Measure** — available line width after left and right margins.

## Angular reference

NASA-STD-3001 Volume 2, Appendix F §F.5.1.7 defines character height as the angle subtended at the eye by an uppercase letter. The current standard requires a minimum of 0.25° and prefers 0.4° or greater. This is a useful safety reference, not a claim that one value is optimal for every reading task, font, contrast, reader, or distance. See [NASA's current Appendix F](https://www.nasa.gov/reference/appendix-f-vol-2/).

For body size `B`, cap-height units `C`, hhea ascent `A`, negative descent `D`, panel density `P`, and eye distance `d`:

```text
vertical units = A - D
cap pixels     = B * C / (A - D)
cap inches     = cap pixels / P
angle degrees  = 2 * atan(cap inches / (2 * d)) * 180 / pi
```

Never substitute nominal point size for cap height; fonts with the same point size can have materially different cap ratios.

## FBInk line geometry

The lab mirrors the integer rounding used by the working FBInk path:

```text
baseline pixels = ceil(B * A / (A - D))
descent pixels  = ceil(B * D / (A - D))
unspaced height = baseline pixels + abs(descent pixels)
gap pixels      = ceil(B * hhea.lineGap / (A - D))
baseline pitch  = unspaced height + gap pixels
```

To request a gap of `G` pixels, the smallest OpenType value that rounds to it is:

```text
minimum hhea.lineGap = floor((G - 1) * (A - D) / B) + 1
```

The exact first physical trial used a 49 px body and 168 line-gap units. That produces an 8 px gap and a 57 px baseline pitch. The minimal equivalent is 161 units; it rounds to the same 8 px result. `reference-recipe.json` keeps 168 to preserve the trial, while new dial exports choose the minimal deterministic value.

## Why absolute pitch matters

Suppose a 52 px face has a 52 px pitch. Reducing body size to 49 px and applying “10% more line spacing” only to the new body would target about 54 px—barely more open than before. ReKindled instead applies the requested pitch change to the old physical baseline: `round(52 * 1.10) = 57 px`, then solves the needed font gap. The two dials remain independent.

## Experiment loop

1. Load the closest preset in `typography-lab/index.html` or call `rekindled_type_lab`.
2. Adjust body, baseline pitch, distance, and measure.
3. Export the JSON recipe.
4. Run `materialize-type-recipe.py` with a legally obtained font.
5. Give the derivative a new family and preset ID.
6. Run the Kindle reader with `-check-layout` on a representative multi-page specimen.
7. Display it on the e-ink panel.
8. Compare at a fixed distance, posture, lighting, and refresh mode.
9. Record comprehension comfort and page density; change one or two dials for the next trial.

The HTML specimen is useful for relationships and gross density. It cannot reproduce FBInk's font engine, hinting, the Kindle panel, e-ink contrast, refresh mode, glare, or the reader's vision. The device is the acceptance test.

## Materializing a recipe

```sh
python3 scripts/materialize-type-recipe.py \
  --source-font /path/to/legal-source.ttf \
  --source-preset kindle/reader/presets/mac/rekindled-mono-air.json \
  --recipe /path/to/exported-recipe.json \
  --output-font build/type/MyDerivedFace-Regular.ttf \
  --output-preset build/type/my-derived-face.json \
  --preset-id my-derived-face \
  --label "My Derived Face / trial 2" \
  --header "ReKindled / trial 2" \
  --derived-family "My Derived Face"
```

The script changes only `hhea.lineGap` and internal naming records, then updates the preset. It does not alter outlines or hinting. Review the source license before modification, local device use, or redistribution.
