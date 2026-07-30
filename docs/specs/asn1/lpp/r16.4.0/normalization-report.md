# Normalization attempt

No normalized module was produced. The immutable original assembly has systematic
DOCX run-boundary loss: normative fields appear as
`transactionIDLPP-TransactionIDOPTIONAL` and `endTransactionBOOLEAN`.
The current provenance extractor concatenates all `w:t` runs without preserving
Word run whitespace. Restoring spaces by splitting identifiers would be
semantic inference and is not an allowed normalization. The safe repair is a
new reviewed extractor that preserves Word run whitespace in a separate
provenance layer; it must not rewrite the existing extraction.
