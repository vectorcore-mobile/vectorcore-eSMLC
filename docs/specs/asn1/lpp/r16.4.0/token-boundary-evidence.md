# Token-boundary evidence

V2 block 002 contains DOCX `w:tab` elements between `transactionID`,
`LPP-TransactionID`, and `OPTIONAL`, and between `endTransaction` and
`BOOLEAN`. V2 renders those explicit XML separators as TAB; no ASN.1 keyword
splitting is used.
