# Minimum-envelope compiler-complete subset

The 19-definition minimum-form envelope closure cannot be an ASN.1 module by itself. `LPP-MessageBody` is a normative CHOICE, so retaining its complete text requires every alternative definition; recursive structural resolution expands to all 646 assignments in `LPP-PDU-Definitions`. The generated compiler target is therefore a byte-identical copy of the normalized PDU module. `LPP-Broadcast-Definitions` is omitted because it is not reachable from `LPP-Message`. No definition, constraint, tag, extension marker, field order, or import was rewritten.
