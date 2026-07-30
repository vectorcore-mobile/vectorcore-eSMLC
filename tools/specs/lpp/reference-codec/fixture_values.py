"""Standards-shaped minimum LPP envelope values for the independent UPER oracle."""

FIXTURES = {
    "minimal-no-transaction": ("LPP-Message", {"endTransaction": False}, "No transaction ID, sequence number, acknowledgement, or body."),
    "transaction-location-server-0": ("LPP-Message", {"transactionID": {"initiator": "locationServer", "transactionNumber": 0}, "endTransaction": False}, "Transaction lower bound and locationServer initiator."),
    "transaction-target-device-255-end": ("LPP-Message", {"transactionID": {"initiator": "targetDevice", "transactionNumber": 255}, "endTransaction": True}, "Transaction upper bound, targetDevice initiator, terminal message."),
    "sequence-zero-ack-requested": ("LPP-Message", {"endTransaction": False, "sequenceNumber": 0, "acknowledgement": {"ackRequested": True}}, "Optional sequence and acknowledgement request."),
    "sequence-255-ack-indicator": ("LPP-Message", {"endTransaction": False, "sequenceNumber": 255, "acknowledgement": {"ackRequested": False, "ackIndicator": 255}}, "Optional sequence upper bound and acknowledgement response form."),
    "request-capabilities-r9-empty": ("LPP-Message", {"endTransaction": False, "lpp-MessageBody": ("c1", ("requestCapabilities", {"criticalExtensions": ("c1", ("requestCapabilities-r9", {}))}))}, "Minimum R9 RequestCapabilities wrapper; all R9 IE members omitted."),
    "provide-capabilities-r9-empty": ("LPP-Message", {"endTransaction": False, "lpp-MessageBody": ("c1", ("provideCapabilities", {"criticalExtensions": ("c1", ("provideCapabilities-r9", {}))}))}, "Minimum R9 ProvideCapabilities wrapper; all R9 IE members omitted."),
    "request-location-information-r9-empty": ("LPP-Message", {"endTransaction": False, "lpp-MessageBody": ("c1", ("requestLocationInformation", {"criticalExtensions": ("c1", ("requestLocationInformation-r9", {}))}))}, "Minimum R9 RequestLocationInformation wrapper; no positioning-method IE supplied."),
    "provide-location-information-r9-empty": ("LPP-Message", {"endTransaction": False, "lpp-MessageBody": ("c1", ("provideLocationInformation", {"criticalExtensions": ("c1", ("provideLocationInformation-r9", {}))}))}, "Minimum R9 ProvideLocationInformation wrapper; no estimate supplied."),
    "abort-r9-empty": ("LPP-Message", {"endTransaction": True, "lpp-MessageBody": ("c1", ("abort", {"criticalExtensions": ("c1", ("abort-r9", {}))}))}, "Minimum R9 Abort wrapper."),
    "error-r9-empty": ("LPP-Message", {"endTransaction": True, "lpp-MessageBody": ("c1", ("error", ("error-r9", {})))}, "Minimum R9 Error wrapper."),
}

NEGATIVE = {
    "transaction-below-range": ("LPP-Message", {"transactionID": {"initiator": "locationServer", "transactionNumber": -1}, "endTransaction": False}),
    "transaction-above-range": ("LPP-Message", {"transactionID": {"initiator": "locationServer", "transactionNumber": 256}, "endTransaction": False}),
    "invalid-initiator": ("LPP-Message", {"transactionID": {"initiator": "invalid", "transactionNumber": 0}, "endTransaction": False}),
    "sequence-above-range": ("LPP-Message", {"endTransaction": False, "sequenceNumber": 256}),
}
