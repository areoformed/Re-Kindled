# Third-party boundary

No third-party executable or font binary is included in this release.

## FBInk

The Kindle reader calls an independently installed `fbink` executable for layout and e-ink refresh. Obtain FBInk from its upstream project and follow its license and device-support instructions. ReKindled does not redistribute it.

## Reference font

The optional reference workflow downloads `PTM55FT.ttf` and its OFL text directly from the Google Fonts repository, verifies a pinned SHA-256 digest, and keeps both under ignored `code/build/source-fonts/`.

PT Mono's OFL declares the Reserved Font Names **PT Sans**, **PT Serif**, **PT Mono**, and **ParaType**. A modified font therefore must use different primary/internal and manufacturer names. `materialize-type-recipe.py` removes those original naming records and defaults to **ReKindled Mono Air** with a local-derivative manufacturer record. Copyright, trademark, description, and OFL records remain for attribution. The original font and OFL remain available from [Google Fonts' PT Mono directory](https://github.com/google/fonts/tree/main/ofl/ptmono).

Do not redistribute the locally materialized derivative without satisfying the OFL, including its license-copy requirements. The same general rule applies to every source font: inspect its license independently. A desktop license or OpenType `fsType` value is not, by itself, blanket permission to modify or redistribute a font.

## Commercial fonts

The materializer accepts any local OpenType font because some owners have valid local licenses. Commercial font names, files, metrics, and paths are intentionally absent from the public package. The operator is responsible for the applicable license.

## fontTools

`fontTools` is an optional build-time dependency used only to rename a local derivative and set its `hhea.lineGap`. It is not required by the MCP server or Kindle reader.

## Go

The MCP server and reader use only the Go standard library. Go itself is not redistributed here.
