# Whitespace-preserving provenance layer

Version 2 reads WordprocessingML in document order. It appends `w:t` text
unchanged, renders `w:tab` as TAB, and renders `w:br`/`w:cr` as newline. It
does not infer ASN.1 token boundaries. The 366 blocks are separate from v1.
