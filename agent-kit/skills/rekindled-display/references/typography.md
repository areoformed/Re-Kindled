# Typography calibration

Treat body pixels, uppercase visual angle, absolute baseline pitch, and measure as separate physical dials.

For body `B`, cap units `C`, hhea ascent `A`, negative descent `D`, PPI `P`, and distance `d`:

```text
cap pixels = B * C / (A - D)
angle      = 2 * atan((cap pixels / P) / (2 * d))
```

Convert radians to degrees. NASA-STD-3001 Volume 2 Appendix F §F.5.1.7 currently uses 0.25° uppercase character height as a minimum and prefers 0.4° or greater. Use this as a reference, not a universal optimum.

The FBInk working path rounds line geometry as:

```text
unspaced = ceil(B*A/(A-D)) + abs(ceil(B*D/(A-D)))
gap px   = ceil(B*hhea.lineGap/(A-D))
pitch    = unspaced + gap px
```

For requested gap `G`, choose the minimal units:

```text
floor((G - 1) * (A - D) / B) + 1
```

This matters when body and pitch change together. Always calculate target pitch from the source preset's old absolute pitch, then solve the new gap. Do not apply a spacing percentage only to the smaller body.

## Agent workflow

1. Call `rekindled_type_lab` with a preset plus either body change or cap-angle target and an independent line-pitch change.
2. Inspect requested versus rounded candidate values.
3. Materialize the recipe to a newly named local font/preset.
4. Preflight representative pages with FBInk.
5. Display on the physical panel at the recorded distance.
6. Photograph trials under equal lighting and record density/comprehension observations.
7. Change one or two dials and repeat.

The reference trial is 49 body pixels, 8 gap pixels, 57 baseline pixels, and approximately 0.325° at 18 inches with the bundled metrics. The exact trial used 168 line-gap units; 161 is the smallest value that rounds to the same 8 pixels.

Never redistribute a source or modified font without license permission. Fonts with Reserved Font Names require new internal primary names for derivatives.
