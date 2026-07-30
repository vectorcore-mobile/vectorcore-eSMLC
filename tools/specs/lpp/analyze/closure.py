#!/usr/bin/env python3
"""Closure calculation for the bounded LPP envelope analysis milestone."""
from __future__ import annotations

POSITIONING_MARKERS = ("A-GNSS", "OTDOA", "ECID", "Sensor", "TBS", "WLAN", "BT", "NR-")
ROOTS = [
    "LPP-Message", "LPP-TransactionID", "Acknowledgement", "RequestCapabilities",
    "ProvideCapabilities", "RequestLocationInformation", "ProvideLocationInformation", "Abort", "Error",
]

# This is a *minimum-form* envelope closure, not a claim that arbitrary LPP
# procedure payloads can be decoded. Each selected R9 IE container has only
# OPTIONAL root payload fields in TS 37.355 V16.4.0, so these definitions are
# sufficient to construct the corresponding empty/minimum wrapper. A later
# compiler-subset and codec phase must add payload dependencies deliberately.
INITIAL_MINIMUM_FORM = [
    "LPP-Message", "LPP-TransactionID", "Initiator", "TransactionNumber",
    "SequenceNumber", "Acknowledgement", "LPP-MessageBody",
    "RequestCapabilities", "RequestCapabilities-r9-IEs",
    "ProvideCapabilities", "ProvideCapabilities-r9-IEs",
    "RequestLocationInformation", "RequestLocationInformation-r9-IEs",
    "ProvideLocationInformation", "ProvideLocationInformation-r9-IEs",
    "Abort", "Abort-r9-IEs", "Error", "Error-r9-IEs",
]


def reachable(graph: dict[str, list[str]], roots: list[str]) -> tuple[list[str], list[str]]:
    seen: set[str] = set()
    unresolved: set[str] = set()
    todo = list(reversed(roots))
    while todo:
        symbol = todo.pop()
        if symbol in seen:
            continue
        seen.add(symbol)
        if symbol not in graph:
            unresolved.add(symbol)
            continue
        todo.extend(reversed(graph[symbol]))
    return sorted(seen), sorted(unresolved)


def classify(symbol: str, root_set: set[str], unresolved: set[str]) -> str:
    if symbol in unresolved:
        return "unresolved"
    if symbol in root_set or symbol in {"SequenceNumber", "Initiator", "TransactionNumber", "LPP-MessageBody"}:
        return "required-core"
    if any(marker in symbol for marker in POSITIONING_MARKERS):
        return "positioning-method"
    if "CommonIE" in symbol or symbol.startswith("EPDU"):
        return "required-common"
    if "-r13" in symbol or "-r16" in symbol:
        return "extension-only"
    return "optional"
