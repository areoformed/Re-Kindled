package main

const instructions = `ReKindled is a private, cable-local Kindle display. Use rekindled_show to place already-paginated text; it rejects overflow before changing the visible page. Use rekindled_status before recovery work, rekindled_presets before choosing type, and rekindled_type_lab to calculate non-destructive sizing and line-pitch candidates. Use rekindled_help for operating guidance. The reader also accepts Kindle touch gestures.`

const manualMarkdown = `# ReKindled MCP

ReKindled gives an agent seven quiet controls over a USB-attached Kindle:

1. **show** — render one or more deliberately paginated text pages.
2. **navigate** — move one page forward or backward.
3. **set_type** — select an installed typography preset.
4. **status** — inspect the connection and reader without changing it.
5. **help** — read this operating guidance by topic.
6. **presets** — list installed type choices and their physical sizing.
7. **type_lab** — calculate body-size, visual-angle, and line-pitch recipes.

Nothing here opens the internet, manages a library, or exposes a general shell.

## A good first call

Call ` + "`rekindled_show`" + ` with a short title and a ` + "`pages`" + ` array. Make page
breaks semantic: end at a paragraph, section, or complete thought. If any page is
too long for the active type and margins, ReKindled returns an error and preserves
the old display.

Example arguments:

` + "```json" + `
{
  "title": "Morning brief",
  "pages": [
    "The important thing first. Two restrained paragraphs belong here.",
    "A second page, complete in itself."
  ],
  "preset": "rekindled-mono-air"
}
` + "```" + `

## Kindle gestures

- Triple-tap the right half: next page.
- Triple-tap the left half: previous page.
- Swipe up: next page.
- Swipe down: previous page.
- Swipe right from the far left edge: exit the reader.

Touch and MCP navigation are peers. A single tap has no display action.

## Typography

` + "`rekindled-mono-air`" + ` is the compact everyday preset. Its 49 px body is 5.8%
smaller than the previous trial. Its 57 px baseline pitch is 9.6% larger than
before, and its cap height is approximately 0.325 degrees at 18 inches. Use
` + "`rekindled_presets`" + ` before selecting another face.

## Resources

- ` + "`rekindled://help`" + ` — this complete manual.
- ` + "`rekindled://presets`" + ` — the local preset catalog as JSON.
- ` + "`rekindled://status`" + ` — live device state as JSON.
- ` + "`rekindled://typography`" + ` — equations and the physical experiment loop.

## Recovery

1. Keep the Kindle connected by USB and leave USBNetLite active.
2. Call ` + "`rekindled_status`" + `; connection errors include the underlying SSH detail.
3. If the reader is stopped, a successful ` + "`rekindled_show`" + ` starts it again.
4. If type selection fails, call ` + "`rekindled_presets`" + ` and use an exact preset id.
5. For layout rejection, split the offending page at a paragraph boundary.

The MCP process writes protocol messages only to stdout. Operational diagnostics
go to stderr. Closing stdin shuts it down cleanly.
`

var helpTopics = map[string]string{
	"overview":   `ReKindled is a cable-local text display. The usual sequence is status (optional), presets (when choosing type), then show. Content is laid out on the Kindle before the visible document is replaced.`,
	"show":       `Send an array of complete pages. Prefer semantic page breaks and modest page lengths. The title is optional and replaces the preset header. A layout error leaves the current display intact.`,
	"type":       `Call rekindled_presets for exact ids, then rekindled_set_type. rekindled-mono-air is the 49 px everyday preset with a 57 px baseline pitch, 9.6 percent more open than the prior trial. Changing type restarts only the reader, not the Kindle.`,
	"typography": `Use rekindled_type_lab for non-destructive calculations. Size can be expressed as a relative body change or an angular cap-height target. Line-pitch change is absolute relative to the current baseline pitch. Materialize the returned hhea line-gap units with scripts/materialize-type-recipe.py, preflight on FBInk, then judge a specimen on the physical panel. Browser previews are geometric aids, not rasterization truth.`,
	"gestures":   `On Kindle: triple-tap right or swipe up for next; triple-tap left or swipe down for previous; swipe right from the far left edge to exit.`,
	"recovery":   `Call rekindled_status first. Keep USBNetLite active. A show call starts a stopped reader. If layout is rejected, shorten or split the named page; if type is rejected, choose an exact id from rekindled_presets.`,
}

const typographyGuide = `# ReKindled typography lab

Treat four values as separate dials: body pixels, angular cap height, absolute
baseline pitch, and margins. A smaller body plus larger line gap can still
produce too little absolute pitch change, so always compare baseline pixels
before and after.

The cap-height angle is derived from the font cap-height ratio, panel PPI, and
viewing distance. FBInk scales OpenType fonts against hhea ascent minus descent,
then adds the hhea line gap between lines. The type lab reports both the requested
change and the rounded device result.

Use this loop:

1. Call rekindled_type_lab or use scripts/type-lab.py.
2. Materialize the recipe into a derived font and preset.
3. Run the reader with -check-layout against a representative specimen.
4. Display the specimen on the physical panel.
5. Photograph before and after at the same distance and lighting.
6. Change one or two dials at a time and record the result.

Never call a browser preview proof of Kindle rendering. Hinting, FBInk layout,
e-ink contrast, refresh mode, viewing distance, and the panel itself all matter.
`
