# TS 37.355 V2 symbol inventory

Two normalized modules contain **652** assignments. There are 0 duplicate definitions and 0 unresolved structural references.

| Module | Symbol | Kind | Line | Direct known references |
|---|---|---|---|---|
| LPP-Broadcast-Definitions | AssistanceDataSIBelement-r15 | type | 12 | CipheringKeyData-r15, SegmentationInfo-r15 |
| LPP-Broadcast-Definitions | CipheringKeyData-r15 | type | 20 |  |
| LPP-Broadcast-Definitions | SegmentationInfo-r15 | type | 25 |  |
| LPP-Broadcast-Definitions | OTDOA-UE-Assisted-r15 | type | 32 | OTDOA-ReferenceCellInfo, OTDOA-NeighbourCellInfoList |
| LPP-Broadcast-Definitions | NR-UEB-TRP-LocationData-r16 | type | 38 | NR-TRP-LocationInfo-r16, NR-DL-PRS-BeamInfo-r16 |
| LPP-Broadcast-Definitions | NR-UEB-TRP-RTD-Info-r16 | type | 44 | NR-RTD-Info-r16 |
| LPP-PDU-Definitions | LPP-Message | type | 7 | LPP-TransactionID, SequenceNumber, Acknowledgement, LPP-MessageBody |
| LPP-PDU-Definitions | SequenceNumber | type | 14 |  |
| LPP-PDU-Definitions | Acknowledgement | type | 15 | SequenceNumber |
| LPP-PDU-Definitions | LPP-MessageBody | type | 20 | RequestCapabilities, ProvideCapabilities, RequestAssistanceData, ProvideAssistanceData, RequestLocationInformation, ProvideLocationInformation, Abort, Error |
| LPP-PDU-Definitions | LPP-TransactionID | type | 36 | Initiator, TransactionNumber |
| LPP-PDU-Definitions | Initiator | type | 41 |  |
| LPP-PDU-Definitions | TransactionNumber | type | 46 |  |
| LPP-PDU-Definitions | RequestCapabilities | type | 48 | RequestCapabilities-r9-IEs |
| LPP-PDU-Definitions | RequestCapabilities-r9-IEs | type | 57 | CommonIEsRequestCapabilities, A-GNSS-RequestCapabilities, OTDOA-RequestCapabilities, ECID-RequestCapabilities, EPDU-Sequence, Sensor-RequestCapabilities-r13, TBS-RequestCapabilities-r13, WLAN-RequestCapabilities-r13, BT-RequestCapabilities-r13, NR-ECID-RequestCapabilities-r16, NR-Multi-RTT-RequestCapabilities-r16, NR-DL-AoD-RequestCapabilities-r16, NR-DL-TDOA-RequestCapabilities-r16, NR-UL-RequestCapabilities-r16 |
| LPP-PDU-Definitions | ProvideCapabilities | type | 81 | ProvideCapabilities-r9-IEs |
| LPP-PDU-Definitions | ProvideCapabilities-r9-IEs | type | 90 | CommonIEsProvideCapabilities, A-GNSS-ProvideCapabilities, OTDOA-ProvideCapabilities, ECID-ProvideCapabilities, EPDU-Sequence, Sensor-ProvideCapabilities-r13, TBS-ProvideCapabilities-r13, WLAN-ProvideCapabilities-r13, BT-ProvideCapabilities-r13, NR-ECID-ProvideCapabilities-r16, NR-Multi-RTT-ProvideCapabilities-r16, NR-DL-AoD-ProvideCapabilities-r16, NR-DL-TDOA-ProvideCapabilities-r16, NR-UL-ProvideCapabilities-r16 |
| LPP-PDU-Definitions | RequestAssistanceData | type | 113 | RequestAssistanceData-r9-IEs |
| LPP-PDU-Definitions | RequestAssistanceData-r9-IEs | type | 122 | CommonIEsRequestAssistanceData, A-GNSS-RequestAssistanceData, OTDOA-RequestAssistanceData, EPDU-Sequence, Sensor-RequestAssistanceData-r14, TBS-RequestAssistanceData-r14, WLAN-RequestAssistanceData-r14, NR-Multi-RTT-RequestAssistanceData-r16, NR-DL-AoD-RequestAssistanceData-r16, NR-DL-TDOA-RequestAssistanceData-r16 |
| LPP-PDU-Definitions | ProvideAssistanceData | type | 139 | ProvideAssistanceData-r9-IEs |
| LPP-PDU-Definitions | ProvideAssistanceData-r9-IEs | type | 148 | CommonIEsProvideAssistanceData, A-GNSS-ProvideAssistanceData, OTDOA-ProvideAssistanceData, EPDU-Sequence, Sensor-ProvideAssistanceData-r14, TBS-ProvideAssistanceData-r14, WLAN-ProvideAssistanceData-r14, NR-Multi-RTT-ProvideAssistanceData-r16, NR-DL-AoD-ProvideAssistanceData-r16, NR-DL-TDOA-ProvideAssistanceData-r16 |
| LPP-PDU-Definitions | RequestLocationInformation | type | 170 | RequestLocationInformation-r9-IEs |
| LPP-PDU-Definitions | RequestLocationInformation-r9-IEs | type | 179 | CommonIEsRequestLocationInformation, A-GNSS-RequestLocationInformation, OTDOA-RequestLocationInformation, ECID-RequestLocationInformation, EPDU-Sequence, Sensor-RequestLocationInformation-r13, TBS-RequestLocationInformation-r13, WLAN-RequestLocationInformation-r13, BT-RequestLocationInformation-r13, NR-ECID-RequestLocationInformation-r16, NR-Multi-RTT-RequestLocationInformation-r16, NR-DL-AoD-RequestLocationInformation-r16, NR-DL-TDOA-RequestLocationInformation-r16 |
| LPP-PDU-Definitions | ProvideLocationInformation | type | 210 | ProvideLocationInformation-r9-IEs |
| LPP-PDU-Definitions | ProvideLocationInformation-r9-IEs | type | 219 | CommonIEsProvideLocationInformation, A-GNSS-ProvideLocationInformation, OTDOA-ProvideLocationInformation, ECID-ProvideLocationInformation, EPDU-Sequence, Sensor-ProvideLocationInformation-r13, TBS-ProvideLocationInformation-r13, WLAN-ProvideLocationInformation-r13, BT-ProvideLocationInformation-r13, NR-ECID-ProvideLocationInformation-r16, NR-Multi-RTT-ProvideLocationInformation-r16, NR-DL-AoD-ProvideLocationInformation-r16, NR-DL-TDOA-ProvideLocationInformation-r16 |
| LPP-PDU-Definitions | Abort | type | 246 | Abort-r9-IEs |
| LPP-PDU-Definitions | Abort-r9-IEs | type | 255 | CommonIEsAbort, EPDU-Sequence |
| LPP-PDU-Definitions | Error | type | 261 | Error-r9-IEs |
| LPP-PDU-Definitions | Error-r9-IEs | type | 265 | CommonIEsError, EPDU-Sequence |
| LPP-PDU-Definitions | AccessTypes | type | 271 |  |
| LPP-PDU-Definitions | ARFCN-ValueEUTRA | type | 280 | maxEARFCN |
| LPP-PDU-Definitions | ARFCN-ValueEUTRA-v9a0 | type | 281 | maxEARFCN-Plus1, maxEARFCN2 |
| LPP-PDU-Definitions | ARFCN-ValueEUTRA-r14 | type | 282 | maxEARFCN2 |
| LPP-PDU-Definitions | ARFCN-ValueNR-r15 | type | 284 |  |
| LPP-PDU-Definitions | ARFCN-ValueUTRA | type | 286 |  |
| LPP-PDU-Definitions | CarrierFreq-NB-r14 | type | 288 | ARFCN-ValueEUTRA-r14, CarrierFreqOffsetNB-r14 |
| LPP-PDU-Definitions | CarrierFreqOffsetNB-r14 | type | 294 |  |
| LPP-PDU-Definitions | CellGlobalIdEUTRA-AndUTRA | type | 299 |  |
| LPP-PDU-Definitions | CellGlobalIdGERAN | type | 311 |  |
| LPP-PDU-Definitions | ECGI | type | 321 |  |
| LPP-PDU-Definitions | Ellipsoid-Point | type | 327 |  |
| LPP-PDU-Definitions | Ellipsoid-PointWithUncertaintyCircle | type | 333 |  |
| LPP-PDU-Definitions | EllipsoidPointWithUncertaintyEllipse | type | 340 |  |
| LPP-PDU-Definitions | EllipsoidPointWithAltitude | type | 350 |  |
| LPP-PDU-Definitions | EllipsoidPointWithAltitudeAndUncertaintyEllipsoid | type | 358 |  |
| LPP-PDU-Definitions | EllipsoidArc | type | 371 |  |
| LPP-PDU-Definitions | EPDU-Sequence | type | 382 | maxEPDU, EPDU |
| LPP-PDU-Definitions | maxEPDU | value | 383 |  |
| LPP-PDU-Definitions | EPDU | type | 384 | EPDU-Identifier, EPDU-Body |
| LPP-PDU-Definitions | EPDU-Identifier | type | 388 | EPDU-ID, EPDU-Name |
| LPP-PDU-Definitions | EPDU-ID | type | 393 |  |
| LPP-PDU-Definitions | EPDU-Name | type | 394 |  |
| LPP-PDU-Definitions | EPDU-Body | type | 395 |  |
| LPP-PDU-Definitions | FreqBandIndicatorNR-r16 | type | 397 |  |
| LPP-PDU-Definitions | HighAccuracyEllipsoidPointWithUncertaintyEllipse-r15 | type | 399 |  |
| LPP-PDU-Definitions | HighAccuracyEllipsoidPointWithAltitudeAndUncertaintyEllipsoid-r15 | type | 408 |  |
| LPP-PDU-Definitions | HorizontalVelocity | type | 420 |  |
| LPP-PDU-Definitions | HorizontalWithVerticalVelocity | type | 425 |  |
| LPP-PDU-Definitions | HorizontalVelocityWithUncertainty | type | 432 |  |
| LPP-PDU-Definitions | HorizontalWithVerticalVelocityAndUncertainty | type | 438 |  |
| LPP-PDU-Definitions | LocationCoordinateTypes | type | 447 |  |
| LPP-PDU-Definitions | NCGI-r15 | type | 462 |  |
| LPP-PDU-Definitions | NR-PhysCellID-r16 | type | 468 |  |
| LPP-PDU-Definitions | PeriodicAssistanceDataControlParameters-r15 | type | 470 | PeriodicSessionID-r15, UpdateCapabilities-r15 |
| LPP-PDU-Definitions | PeriodicSessionID-r15 | type | 477 |  |
| LPP-PDU-Definitions | UpdateCapabilities-r15 | type | 482 |  |
| LPP-PDU-Definitions | Polygon | type | 484 | PolygonPoints |
| LPP-PDU-Definitions | PolygonPoints | type | 485 |  |
| LPP-PDU-Definitions | PositioningModes | type | 491 |  |
| LPP-PDU-Definitions | SegmentationInfo-r14 | type | 499 |  |
| LPP-PDU-Definitions | VelocityTypes | type | 501 |  |
| LPP-PDU-Definitions | CommonIEsRequestCapabilities | type | 509 |  |
| LPP-PDU-Definitions | CommonIEsProvideCapabilities | type | 517 | SegmentationInfo-r14 |
| LPP-PDU-Definitions | CommonIEsRequestAssistanceData | type | 526 | ECGI, SegmentationInfo-r14, PeriodicAssistanceDataControlParameters-r15, NCGI-r15 |
| LPP-PDU-Definitions | CommonIEsProvideAssistanceData | type | 540 | SegmentationInfo-r14, PeriodicAssistanceDataControlParameters-r15 |
| LPP-PDU-Definitions | CommonIEsRequestLocationInformation | type | 551 | LocationInformationType, TriggeredReportingCriteria, PeriodicalReportingCriteria, AdditionalInformation, QoS, Environment, LocationCoordinateTypes, VelocityTypes, MessageSizeLimitNB-r14, SegmentationInfo-r14 |
| LPP-PDU-Definitions | LocationInformationType | type | 568 |  |
| LPP-PDU-Definitions | PeriodicalReportingCriteria | type | 575 |  |
| LPP-PDU-Definitions | TriggeredReportingCriteria | type | 585 | ReportingDuration |
| LPP-PDU-Definitions | ReportingDuration | type | 590 |  |
| LPP-PDU-Definitions | AdditionalInformation | type | 591 |  |
| LPP-PDU-Definitions | QoS | type | 596 | HorizontalAccuracy, VerticalAccuracy, ResponseTime, ResponseTimeNB-r14, HorizontalAccuracyExt-r15, VerticalAccuracyExt-r15 |
| LPP-PDU-Definitions | HorizontalAccuracy | type | 609 |  |
| LPP-PDU-Definitions | VerticalAccuracy | type | 614 |  |
| LPP-PDU-Definitions | HorizontalAccuracyExt-r15 | type | 619 |  |
| LPP-PDU-Definitions | VerticalAccuracyExt-r15 | type | 624 |  |
| LPP-PDU-Definitions | ResponseTime | type | 629 |  |
| LPP-PDU-Definitions | ResponseTimeNB-r14 | type | 637 |  |
| LPP-PDU-Definitions | Environment | type | 644 |  |
| LPP-PDU-Definitions | MessageSizeLimitNB-r14 | type | 650 |  |
| LPP-PDU-Definitions | CommonIEsProvideLocationInformation | type | 655 | LocationCoordinates, Velocity, LocationError, EarlyFixReport-r12, LocationSource-r13, SegmentationInfo-r14 |
| LPP-PDU-Definitions | LocationCoordinates | type | 669 | Ellipsoid-Point, Ellipsoid-PointWithUncertaintyCircle, EllipsoidPointWithUncertaintyEllipse, Polygon, EllipsoidPointWithAltitude, EllipsoidPointWithAltitudeAndUncertaintyEllipsoid, EllipsoidArc, HighAccuracyEllipsoidPointWithUncertaintyEllipse-r15, HighAccuracyEllipsoidPointWithAltitudeAndUncertaintyEllipsoid-r15 |
| LPP-PDU-Definitions | Velocity | type | 684 | HorizontalVelocity, HorizontalWithVerticalVelocity, HorizontalVelocityWithUncertainty, HorizontalWithVerticalVelocityAndUncertainty |
| LPP-PDU-Definitions | LocationError | type | 692 | LocationFailureCause |
| LPP-PDU-Definitions | LocationFailureCause | type | 696 |  |
| LPP-PDU-Definitions | EarlyFixReport-r12 | type | 703 |  |
| LPP-PDU-Definitions | LocationSource-r13 | type | 707 |  |
| LPP-PDU-Definitions | CommonIEsAbort | type | 717 |  |
| LPP-PDU-Definitions | CommonIEsError | type | 728 |  |
| LPP-PDU-Definitions | DL-PRS-ID-Info-r16 | type | 740 | nrMaxResourceIDs-r16, NR-DL-PRS-ResourceID-r16, NR-DL-PRS-ResourceSetID-r16 |
| LPP-PDU-Definitions | NR-AdditionalPathList-r16 | type | 747 | NR-AdditionalPath-r16 |
| LPP-PDU-Definitions | NR-AdditionalPath-r16 | type | 748 | NR-TimingQuality-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-AssistanceData-r16 | type | 762 | DL-PRS-ID-Info-r16, nrMaxFreqLayers-r16, NR-DL-PRS-AssistanceDataPerFreq-r16, nrMaxTRPs-r16, NR-SSB-Config-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-AssistanceDataPerFreq-r16 | type | 770 | NR-DL-PRS-PositioningFrequencyLayer-r16, nrMaxTRPsPerFreq-r16, NR-DL-PRS-AssistanceDataPerTRP-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-AssistanceDataPerTRP-r16 | type | 777 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15, NR-DL-PRS-SFN0-Offset-r16, NR-DL-PRS-Info-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-PositioningFrequencyLayer-r16 | type | 789 | ARFCN-ValueNR-r15 |
| LPP-PDU-Definitions | NR-DL-PRS-SFN0-Offset-r16 | type | 798 |  |
| LPP-PDU-Definitions | NR-DL-PRS-BeamInfo-r16 | type | 803 | nrMaxFreqLayers-r16, NR-DL-PRS-BeamInfoPerFreqLayer-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-BeamInfoPerFreqLayer-r16 | type | 805 | nrMaxTRPsPerFreq-r16, NR-DL-PRS-BeamInfoPerTRP-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-BeamInfoPerTRP-r16 | type | 807 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15, LCS-GCS-TranslationParameter-r16, DL-PRS-BeamInfoSet-r16 |
| LPP-PDU-Definitions | DL-PRS-BeamInfoSet-r16 | type | 818 | nrMaxSetsPerTrp-r16, DL-PRS-BeamInfoResourceSet-r16 |
| LPP-PDU-Definitions | DL-PRS-BeamInfoResourceSet-r16 | type | 820 | nrMaxResourcesPerSet-r16, DL-PRS-BeamInfoElement-r16 |
| LPP-PDU-Definitions | DL-PRS-BeamInfoElement-r16 | type | 822 |  |
| LPP-PDU-Definitions | LCS-GCS-TranslationParameter-r16 | type | 829 |  |
| LPP-PDU-Definitions | NR-DL-PRS-Info-r16 | type | 839 | nrMaxSetsPerTrp-r16, NR-DL-PRS-ResourceSet-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-ResourceSet-r16 | type | 844 | NR-DL-PRS-ResourceSetID-r16, NR-DL-PRS-Periodicity-and-ResourceSetSlotOffset-r16, DL-PRS-MutingOption1-r16, DL-PRS-MutingOption2-r16, nrMaxResourcesPerSet-r16, NR-DL-PRS-Resource-r16 |
| LPP-PDU-Definitions | DL-PRS-MutingOption1-r16 | type | 860 | NR-MutingPattern-r16 |
| LPP-PDU-Definitions | DL-PRS-MutingOption2-r16 | type | 866 | NR-MutingPattern-r16 |
| LPP-PDU-Definitions | NR-MutingPattern-r16 | type | 870 |  |
| LPP-PDU-Definitions | NR-DL-PRS-Resource-r16 | type | 879 | NR-DL-PRS-ResourceID-r16, nrMaxResourceOffsetValue-1-r16, DL-PRS-QCL-Info-r16 |
| LPP-PDU-Definitions | DL-PRS-QCL-Info-r16 | type | 894 | NR-PhysCellID-r16, NR-DL-PRS-ResourceID-r16, NR-DL-PRS-ResourceSetID-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-Periodicity-and-ResourceSetSlotOffset-r16 | type | 905 |  |
| LPP-PDU-Definitions | NR-DL-PRS-ProcessingCapability-r16 | type | 989 | nrMaxBands-r16, PRS-ProcessingCapabilityPerBand-r16 |
| LPP-PDU-Definitions | PRS-ProcessingCapabilityPerBand-r16 | type | 996 | FreqBandIndicatorNR-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-QCL-ProcessingCapability-r16 | type | 1028 | nrMaxBands-r16, DL-PRS-QCL-ProcessingCapabilityPerBand-r16 |
| LPP-PDU-Definitions | DL-PRS-QCL-ProcessingCapabilityPerBand-r16 | type | 1033 | FreqBandIndicatorNR-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-ResourceID-r16 | type | 1040 | nrMaxNumDL-PRS-ResourcesPerSet-1-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-ResourcesCapability-r16 | type | 1042 | nrMaxBands-r16, DL-PRS-ResourcesCapabilityPerBand-r16, DL-PRS-ResourcesBandCombinationList-r16 |
| LPP-PDU-Definitions | DL-PRS-ResourcesCapabilityPerBand-r16 | type | 1053 | FreqBandIndicatorNR-r16 |
| LPP-PDU-Definitions | DL-PRS-ResourcesBandCombinationList-r16 | type | 1061 | maxBandComb-r16, DL-PRS-ResourcesBandCombination-r16 |
| LPP-PDU-Definitions | DL-PRS-ResourcesBandCombination-r16 | type | 1063 | maxSimultaneousBands-r16, FreqBandIndicatorNR-r16 |
| LPP-PDU-Definitions | NR-DL-PRS-ResourceSetID-r16 | type | 1084 | nrMaxNumDL-PRS-ResourceSetsPerTRP-1-r16 |
| LPP-PDU-Definitions | NR-PositionCalculationAssistance-r16 | type | 1086 | NR-TRP-LocationInfo-r16, NR-DL-PRS-BeamInfo-r16, NR-RTD-Info-r16 |
| LPP-PDU-Definitions | NR-RTD-Info-r16 | type | 1093 | ReferenceTRP-RTD-Info-r16, RTD-InfoList-r16 |
| LPP-PDU-Definitions | ReferenceTRP-RTD-Info-r16 | type | 1098 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15, NR-TimingQuality-r16 |
| LPP-PDU-Definitions | RTD-InfoList-r16 | type | 1111 | nrMaxFreqLayers-r16, RTD-InfoListPerFreqLayer-r16 |
| LPP-PDU-Definitions | RTD-InfoListPerFreqLayer-r16 | type | 1112 | nrMaxTRPsPerFreq-r16, RTD-InfoElement-r16 |
| LPP-PDU-Definitions | RTD-InfoElement-r16 | type | 1113 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15, NR-TimingQuality-r16 |
| LPP-PDU-Definitions | NR-SelectedDL-PRS-IndexList-r16 | type | 1123 | nrMaxFreqLayers-r16, NR-SelectedDL-PRS-PerFreq-r16 |
| LPP-PDU-Definitions | NR-SelectedDL-PRS-PerFreq-r16 | type | 1125 | nrMaxFreqLayers-1-r16, nrMaxTRPsPerFreq-r16, NR-SelectedDL-PRS-IndexPerTRP-r16 |
| LPP-PDU-Definitions | NR-SelectedDL-PRS-IndexPerTRP-r16 | type | 1132 | nrMaxTRPsPerFreq-1-r16, nrMaxSetsPerTrp-r16, DL-SelectedPRS-ResourceSetIndex-r16 |
| LPP-PDU-Definitions | DL-SelectedPRS-ResourceSetIndex-r16 | type | 1139 | nrMaxSetsPerTrp-1-r16, nrMaxResourcesPerSet-r16, DL-SelectedPRS-ResourceIndex-r16 |
| LPP-PDU-Definitions | DL-SelectedPRS-ResourceIndex-r16 | type | 1145 | nrMaxNumDL-PRS-ResourcesPerSet-1-r16 |
| LPP-PDU-Definitions | NR-SSB-Config-r16 | type | 1150 | NR-PhysCellID-r16, ARFCN-ValueNR-r15 |
| LPP-PDU-Definitions | NR-TimeStamp-r16 | type | 1166 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15 |
| LPP-PDU-Definitions | NR-TimingQuality-r16 | type | 1181 |  |
| LPP-PDU-Definitions | NR-TRP-LocationInfo-r16 | type | 1187 | nrMaxFreqLayers-r16, NR-TRP-LocationInfoPerFreqLayer-r16 |
| LPP-PDU-Definitions | NR-TRP-LocationInfoPerFreqLayer-r16 | type | 1189 | ReferencePoint-r16, nrMaxTRPsPerFreq-r16, TRP-LocationInfoElement-r16 |
| LPP-PDU-Definitions | TRP-LocationInfoElement-r16 | type | 1195 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15, RelativeLocation-r16, nrMaxSetsPerTrp-r16, DL-PRS-ResourceSets-TRP-Element-r16 |
| LPP-PDU-Definitions | DL-PRS-ResourceSets-TRP-Element-r16 | type | 1206 | RelativeLocation-r16, nrMaxResourcesPerSet-r16, DL-PRS-Resource-ARP-Element-r16 |
| LPP-PDU-Definitions | DL-PRS-Resource-ARP-Element-r16 | type | 1212 | RelativeLocation-r16 |
| LPP-PDU-Definitions | NR-UL-SRS-Capability-r16 | type | 1217 | nrMaxBands-r16, SRS-CapabilityPerBand-r16, nrMaxConfiguredBands-r16, SRS-PosResourcesPerBand-r16 |
| LPP-PDU-Definitions | SRS-CapabilityPerBand-r16 | type | 1228 | FreqBandIndicatorNR-r16, OLPC-SRS-Pos-r16, SpatialRelationsSRS-Pos-r16 |
| LPP-PDU-Definitions | OLPC-SRS-Pos-r16 | type | 1234 |  |
| LPP-PDU-Definitions | SpatialRelationsSRS-Pos-r16 | type | 1241 |  |
| LPP-PDU-Definitions | SRS-PosResourcesPerBand-r16 | type | 1250 | FreqBandIndicatorNR-r16 |
| LPP-PDU-Definitions | ReferencePoint-r16 | type | 1262 | EllipsoidPointWithAltitudeAndUncertaintyEllipsoid, HighAccuracyEllipsoidPointWithAltitudeAndUncertaintyEllipsoid-r15 |
| LPP-PDU-Definitions | RelativeLocation-r16 | type | 1271 | Delta-Latitude-r16, Delta-Longitude-r16, Delta-Height-r16, LocationUncertainty-r16 |
| LPP-PDU-Definitions | Delta-Latitude-r16 | type | 1280 |  |
| LPP-PDU-Definitions | Delta-Longitude-r16 | type | 1285 |  |
| LPP-PDU-Definitions | Delta-Height-r16 | type | 1290 |  |
| LPP-PDU-Definitions | LocationUncertainty-r16 | type | 1295 |  |
| LPP-PDU-Definitions | OTDOA-ProvideAssistanceData | type | 1302 | OTDOA-ReferenceCellInfo, OTDOA-NeighbourCellInfoList, OTDOA-Error, OTDOA-ReferenceCellInfoNB-r14, OTDOA-NeighbourCellInfoListNB-r14 |
| LPP-PDU-Definitions | OTDOA-ReferenceCellInfo | type | 1313 | ECGI, ARFCN-ValueEUTRA, PRS-Info, ARFCN-ValueEUTRA-v9a0, maxAddPRSconfig-r14, TDD-Config-v1520 |
| LPP-PDU-Definitions | maxAddPRSconfig-r14 | value | 1341 |  |
| LPP-PDU-Definitions | PRS-Info | type | 1343 | maxAvailNarrowBands-Minus1-r14 |
| LPP-PDU-Definitions | maxAvailNarrowBands-Minus1-r14 | value | 1372 |  |
| LPP-PDU-Definitions | TDD-Config-v1520 | type | 1374 |  |
| LPP-PDU-Definitions | OTDOA-NeighbourCellInfoList | type | 1379 | maxFreqLayers, OTDOA-NeighbourFreqInfo |
| LPP-PDU-Definitions | OTDOA-NeighbourFreqInfo | type | 1380 | OTDOA-NeighbourCellInfoElement |
| LPP-PDU-Definitions | OTDOA-NeighbourCellInfoElement | type | 1381 | ECGI, ARFCN-ValueEUTRA, PRS-Info, ARFCN-ValueEUTRA-v9a0, maxAddPRSconfig-r14, Add-PRSconfigNeighbourElement-r14, TDD-Config-v1520 |
| LPP-PDU-Definitions | Add-PRSconfigNeighbourElement-r14 | type | 1412 | PRS-Info |
| LPP-PDU-Definitions | maxFreqLayers | value | 1416 |  |
| LPP-PDU-Definitions | OTDOA-ReferenceCellInfoNB-r14 | type | 1418 | ECGI, CarrierFreq-NB-r14, ARFCN-ValueEUTRA-r14, PRS-Info-NB-r14, TDD-Config-v1520 |
| LPP-PDU-Definitions | PRS-Info-NB-r14 | type | 1435 | maxCarrier-r14, NPRS-Info-r14 |
| LPP-PDU-Definitions | NPRS-Info-r14 | type | 1436 | CarrierFreq-NB-r14 |
| LPP-PDU-Definitions | maxCarrier-r14 | value | 1492 |  |
| LPP-PDU-Definitions | OTDOA-NeighbourCellInfoListNB-r14 | type | 1494 | maxCells-r14, OTDOA-NeighbourCellInfoNB-r14 |
| LPP-PDU-Definitions | OTDOA-NeighbourCellInfoNB-r14 | type | 1496 | ECGI, CarrierFreq-NB-r14, ARFCN-ValueEUTRA-r14, PRS-Info-NB-r14, TDD-Config-v1520 |
| LPP-PDU-Definitions | maxCells-r14 | value | 1519 |  |
| LPP-PDU-Definitions | OTDOA-RequestAssistanceData | type | 1521 |  |
| LPP-PDU-Definitions | OTDOA-ProvideLocationInformation | type | 1532 | OTDOA-SignalMeasurementInformation, OTDOA-Error, OTDOA-SignalMeasurementInformation-NB-r14 |
| LPP-PDU-Definitions | OTDOA-SignalMeasurementInformation | type | 1542 | ECGI, ARFCN-ValueEUTRA, OTDOA-MeasQuality, NeighbourMeasurementList, ARFCN-ValueEUTRA-v9a0, AdditionalPathList-r14, CarrierFreqOffsetNB-r14, MotionTimeSource-r15 |
| LPP-PDU-Definitions | NeighbourMeasurementList | type | 1565 | NeighbourMeasurementElement |
| LPP-PDU-Definitions | NeighbourMeasurementElement | type | 1566 | ECGI, ARFCN-ValueEUTRA, OTDOA-MeasQuality, ARFCN-ValueEUTRA-v9a0, AdditionalPathList-r14, CarrierFreqOffsetNB-r14 |
| LPP-PDU-Definitions | AdditionalPathList-r14 | type | 1588 | maxPaths-r14, AdditionalPath-r14 |
| LPP-PDU-Definitions | maxPaths-r14 | value | 1589 |  |
| LPP-PDU-Definitions | MotionTimeSource-r15 | type | 1590 |  |
| LPP-PDU-Definitions | OTDOA-SignalMeasurementInformation-NB-r14 | type | 1595 | ECGI, ARFCN-ValueEUTRA-r14, OTDOA-MeasQuality, NeighbourMeasurementList-NB-r14, AdditionalPathList-r14, CarrierFreqOffsetNB-r14 |
| LPP-PDU-Definitions | NeighbourMeasurementList-NB-r14 | type | 1610 | NeighbourMeasurementElement-NB-r14 |
| LPP-PDU-Definitions | NeighbourMeasurementElement-NB-r14 | type | 1611 | ECGI, ARFCN-ValueEUTRA-r14, OTDOA-MeasQuality, AdditionalPathList-r14, CarrierFreqOffsetNB-r14 |
| LPP-PDU-Definitions | OTDOA-MeasQuality | type | 1628 |  |
| LPP-PDU-Definitions | AdditionalPath-r14 | type | 1635 | OTDOA-MeasQuality |
| LPP-PDU-Definitions | OTDOA-RequestLocationInformation | type | 1641 |  |
| LPP-PDU-Definitions | OTDOA-ProvideCapabilities | type | 1653 | maxBands, SupportedBandEUTRA, SupportedBandEUTRA-v9a0 |
| LPP-PDU-Definitions | maxBands | value | 1680 |  |
| LPP-PDU-Definitions | SupportedBandEUTRA | type | 1681 | maxFBI |
| LPP-PDU-Definitions | SupportedBandEUTRA-v9a0 | type | 1684 | maxFBI-Plus1, maxFBI2 |
| LPP-PDU-Definitions | maxFBI | value | 1687 |  |
| LPP-PDU-Definitions | maxFBI-Plus1 | value | 1688 |  |
| LPP-PDU-Definitions | maxFBI2 | value | 1689 |  |
| LPP-PDU-Definitions | OTDOA-RequestCapabilities | type | 1691 |  |
| LPP-PDU-Definitions | OTDOA-Error | type | 1695 | OTDOA-LocationServerErrorCauses, OTDOA-TargetDeviceErrorCauses |
| LPP-PDU-Definitions | OTDOA-LocationServerErrorCauses | type | 1701 |  |
| LPP-PDU-Definitions | OTDOA-TargetDeviceErrorCauses | type | 1710 |  |
| LPP-PDU-Definitions | A-GNSS-ProvideAssistanceData | type | 1721 | GNSS-CommonAssistData, GNSS-GenericAssistData, A-GNSS-Error, GNSS-PeriodicAssistData-r15 |
| LPP-PDU-Definitions | GNSS-CommonAssistData | type | 1731 | GNSS-ReferenceTime, GNSS-ReferenceLocation, GNSS-IonosphericModel, GNSS-EarthOrientationParameters, GNSS-RTK-ReferenceStationInfo-r15, GNSS-RTK-CommonObservationInfo-r15, GNSS-RTK-AuxiliaryStationData-r15, GNSS-SSR-CorrectionPoints-r16 |
| LPP-PDU-Definitions | GNSS-GenericAssistData | type | 1751 | GNSS-GenericAssistDataElement |
| LPP-PDU-Definitions | GNSS-GenericAssistDataElement | type | 1752 | GNSS-ID, SBAS-ID, GNSS-TimeModelList, GNSS-DifferentialCorrections, GNSS-NavigationModel, GNSS-RealTimeIntegrity, GNSS-DataBitAssistance, GNSS-AcquisitionAssistance, GNSS-Almanac, GNSS-UTC-Model, GNSS-AuxiliaryInformation, BDS-DifferentialCorrections-r12, BDS-GridModelParameter-r12, GNSS-RTK-Observations-r15, GLO-RTK-BiasInformation-r15, GNSS-RTK-MAC-CorrectionDifferences-r15, GNSS-RTK-Residuals-r15, GNSS-RTK-FKP-Gradients-r15, GNSS-SSR-OrbitCorrections-r15, GNSS-SSR-ClockCorrections-r15, GNSS-SSR-CodeBias-r15, GNSS-SSR-URA-r16, GNSS-SSR-PhaseBias-r16, GNSS-SSR-STEC-Correction-r16, GNSS-SSR-GriddedCorrection-r16, NavIC-DifferentialCorrections-r16, NavIC-GridModelParameter-r16 |
| LPP-PDU-Definitions | GNSS-PeriodicAssistData-r15 | type | 1798 | GNSS-PeriodicControlParam-r15 |
| LPP-PDU-Definitions | GNSS-ReferenceTime | type | 1819 | GNSS-SystemTime, GNSS-ReferenceTimeForOneCell |
| LPP-PDU-Definitions | GNSS-ReferenceTimeForOneCell | type | 1826 | NetworkTime |
| LPP-PDU-Definitions | GNSS-SystemTime | type | 1833 | GNSS-ID, GPS-TOW-Assist |
| LPP-PDU-Definitions | GPS-TOW-Assist | type | 1843 | GPS-TOW-AssistElement |
| LPP-PDU-Definitions | GPS-TOW-AssistElement | type | 1844 |  |
| LPP-PDU-Definitions | NetworkTime | type | 1853 | CellGlobalIdEUTRA-AndUTRA, ARFCN-ValueEUTRA, ARFCN-ValueEUTRA-v9a0, ARFCN-ValueUTRA, CellGlobalIdGERAN, ECGI, CarrierFreq-NB-r14, NCGI-r15, ARFCN-ValueNR-r15 |
| LPP-PDU-Definitions | GNSS-ReferenceLocation | type | 1904 | EllipsoidPointWithAltitudeAndUncertaintyEllipsoid |
| LPP-PDU-Definitions | GNSS-IonosphericModel | type | 1909 | KlobucharModelParameter, NeQuickModelParameter, KlobucharModel2Parameter-r16 |
| LPP-PDU-Definitions | KlobucharModelParameter | type | 1917 |  |
| LPP-PDU-Definitions | KlobucharModel2Parameter-r16 | type | 1930 |  |
| LPP-PDU-Definitions | NeQuickModelParameter | type | 1943 |  |
| LPP-PDU-Definitions | GNSS-EarthOrientationParameters | type | 1955 |  |
| LPP-PDU-Definitions | GNSS-RTK-ReferenceStationInfo-r15 | type | 1966 | GNSS-ReferenceStationID-r15, AntennaDescription-r15, AntennaReferencePointUnc-r15, PhysicalReferenceStationInfo-r15, EqualIntegerAmbiguityLevel-r16 |
| LPP-PDU-Definitions | AntennaDescription-r15 | type | 1981 |  |
| LPP-PDU-Definitions | AntennaReferencePointUnc-r15 | type | 1986 |  |
| LPP-PDU-Definitions | PhysicalReferenceStationInfo-r15 | type | 1995 | GNSS-ReferenceStationID-r15, AntennaReferencePointUnc-r15 |
| LPP-PDU-Definitions | EqualIntegerAmbiguityLevel-r16 | type | 2003 | ReferenceStationList-r16 |
| LPP-PDU-Definitions | ReferenceStationList-r16 | type | 2007 | GNSS-ReferenceStationID-r15 |
| LPP-PDU-Definitions | GNSS-RTK-CommonObservationInfo-r15 | type | 2009 | GNSS-ReferenceStationID-r15 |
| LPP-PDU-Definitions | GNSS-RTK-AuxiliaryStationData-r15 | type | 2018 | GNSS-NetworkID-r15, GNSS-SubNetworkID-r15, GNSS-ReferenceStationID-r15, AuxiliaryStationList-r15 |
| LPP-PDU-Definitions | AuxiliaryStationList-r15 | type | 2025 | AuxiliaryStationElement-r15 |
| LPP-PDU-Definitions | AuxiliaryStationElement-r15 | type | 2026 | GNSS-ReferenceStationID-r15, Aux-ARP-Unc-r15 |
| LPP-PDU-Definitions | Aux-ARP-Unc-r15 | type | 2034 |  |
| LPP-PDU-Definitions | GNSS-SSR-CorrectionPoints-r16 | type | 2042 | GNSS-SSR-ListOfCorrectionPoints-r16, GNSS-SSR-ArrayOfCorrectionPoints-r16 |
| LPP-PDU-Definitions | GNSS-SSR-ListOfCorrectionPoints-r16 | type | 2050 | RelativeLocationElement-r16 |
| LPP-PDU-Definitions | RelativeLocationElement-r16 | type | 2056 |  |
| LPP-PDU-Definitions | GNSS-SSR-ArrayOfCorrectionPoints-r16 | type | 2061 |  |
| LPP-PDU-Definitions | GNSS-TimeModelList | type | 2072 | GNSS-TimeModelElement |
| LPP-PDU-Definitions | GNSS-TimeModelElement | type | 2073 |  |
| LPP-PDU-Definitions | GNSS-DifferentialCorrections | type | 2084 | DGNSS-SgnTypeList |
| LPP-PDU-Definitions | DGNSS-SgnTypeList | type | 2089 | DGNSS-SgnTypeElement |
| LPP-PDU-Definitions | DGNSS-SgnTypeElement | type | 2090 | GNSS-SignalID, DGNSS-SatList |
| LPP-PDU-Definitions | DGNSS-SatList | type | 2096 | DGNSS-CorrectionsElement |
| LPP-PDU-Definitions | DGNSS-CorrectionsElement | type | 2097 | SV-ID |
| LPP-PDU-Definitions | GNSS-NavigationModel | type | 2108 | GNSS-NavModelSatelliteList |
| LPP-PDU-Definitions | GNSS-NavModelSatelliteList | type | 2113 | GNSS-NavModelSatelliteElement |
| LPP-PDU-Definitions | GNSS-NavModelSatelliteElement | type | 2114 | SV-ID, GNSS-ClockModel, GNSS-OrbitModel |
| LPP-PDU-Definitions | GNSS-ClockModel | type | 2124 | StandardClockModelList, NAV-ClockModel, CNAV-ClockModel, GLONASS-ClockModel, SBAS-ClockModel, BDS-ClockModel-r12, BDS-ClockModel2-r16, NavIC-ClockModel-r16 |
| LPP-PDU-Definitions | GNSS-OrbitModel | type | 2135 | NavModelKeplerianSet, NavModelNAV-KeplerianSet, NavModelCNAV-KeplerianSet, NavModel-GLONASS-ECEF, NavModel-SBAS-ECEF, NavModel-BDS-KeplerianSet-r12, NavModel-BDS-KeplerianSet2-r16, NavModel-NavIC-KeplerianSet-r16 |
| LPP-PDU-Definitions | StandardClockModelList | type | 2147 | StandardClockModelElement |
| LPP-PDU-Definitions | StandardClockModelElement | type | 2148 |  |
| LPP-PDU-Definitions | NAV-ClockModel | type | 2159 |  |
| LPP-PDU-Definitions | CNAV-ClockModel | type | 2168 |  |
| LPP-PDU-Definitions | GLONASS-ClockModel | type | 2187 |  |
| LPP-PDU-Definitions | SBAS-ClockModel | type | 2194 |  |
| LPP-PDU-Definitions | BDS-ClockModel-r12 | type | 2201 |  |
| LPP-PDU-Definitions | BDS-ClockModel2-r16 | type | 2211 |  |
| LPP-PDU-Definitions | NavIC-ClockModel-r16 | type | 2221 |  |
| LPP-PDU-Definitions | NavModelKeplerianSet | type | 2230 |  |
| LPP-PDU-Definitions | NavModelNAV-KeplerianSet | type | 2250 |  |
| LPP-PDU-Definitions | NavModelCNAV-KeplerianSet | type | 2283 |  |
| LPP-PDU-Definitions | NavModel-GLONASS-ECEF | type | 2306 |  |
| LPP-PDU-Definitions | NavModel-SBAS-ECEF | type | 2323 |  |
| LPP-PDU-Definitions | NavModel-BDS-KeplerianSet-r12 | type | 2338 |  |
| LPP-PDU-Definitions | NavModel-BDS-KeplerianSet2-r16 | type | 2360 |  |
| LPP-PDU-Definitions | NavModel-NavIC-KeplerianSet-r16 | type | 2383 |  |
| LPP-PDU-Definitions | GNSS-RealTimeIntegrity | type | 2404 | GNSS-BadSignalList |
| LPP-PDU-Definitions | GNSS-BadSignalList | type | 2408 | BadSignalElement |
| LPP-PDU-Definitions | BadSignalElement | type | 2409 | SV-ID, GNSS-SignalIDs |
| LPP-PDU-Definitions | GNSS-DataBitAssistance | type | 2415 | GNSS-DataBitsSatList |
| LPP-PDU-Definitions | GNSS-DataBitsSatList | type | 2421 | GNSS-DataBitsSatElement |
| LPP-PDU-Definitions | GNSS-DataBitsSatElement | type | 2422 | SV-ID, GNSS-DataBitsSgnList |
| LPP-PDU-Definitions | GNSS-DataBitsSgnList | type | 2427 | GNSS-DataBitsSgnElement |
| LPP-PDU-Definitions | GNSS-DataBitsSgnElement | type | 2428 | GNSS-SignalID |
| LPP-PDU-Definitions | GNSS-AcquisitionAssistance | type | 2434 | GNSS-SignalID, GNSS-AcquisitionAssistList |
| LPP-PDU-Definitions | GNSS-AcquisitionAssistList | type | 2440 | GNSS-AcquisitionAssistElement |
| LPP-PDU-Definitions | GNSS-AcquisitionAssistElement | type | 2441 | SV-ID |
| LPP-PDU-Definitions | GNSS-Almanac | type | 2460 | GNSS-AlmanacList |
| LPP-PDU-Definitions | GNSS-AlmanacList | type | 2475 | GNSS-AlmanacElement |
| LPP-PDU-Definitions | GNSS-AlmanacElement | type | 2476 | AlmanacKeplerianSet, AlmanacNAV-KeplerianSet, AlmanacReducedKeplerianSet, AlmanacMidiAlmanacSet, AlmanacGLONASS-AlmanacSet, AlmanacECEF-SBAS-AlmanacSet, AlmanacBDS-AlmanacSet-r12, AlmanacNavIC-AlmanacSet-r16 |
| LPP-PDU-Definitions | AlmanacKeplerianSet | type | 2488 | SV-ID |
| LPP-PDU-Definitions | AlmanacNAV-KeplerianSet | type | 2504 | SV-ID |
| LPP-PDU-Definitions | AlmanacReducedKeplerianSet | type | 2519 | SV-ID |
| LPP-PDU-Definitions | AlmanacMidiAlmanacSet | type | 2530 | SV-ID |
| LPP-PDU-Definitions | AlmanacGLONASS-AlmanacSet | type | 2547 |  |
| LPP-PDU-Definitions | AlmanacECEF-SBAS-AlmanacSet | type | 2564 | SV-ID |
| LPP-PDU-Definitions | AlmanacBDS-AlmanacSet-r12 | type | 2578 | SV-ID |
| LPP-PDU-Definitions | AlmanacNavIC-AlmanacSet-r16 | type | 2594 | SV-ID |
| LPP-PDU-Definitions | GNSS-UTC-Model | type | 2608 | UTC-ModelSet1, UTC-ModelSet2, UTC-ModelSet3, UTC-ModelSet4, UTC-ModelSet5-r12 |
| LPP-PDU-Definitions | UTC-ModelSet1 | type | 2617 |  |
| LPP-PDU-Definitions | UTC-ModelSet2 | type | 2629 |  |
| LPP-PDU-Definitions | UTC-ModelSet3 | type | 2645 |  |
| LPP-PDU-Definitions | UTC-ModelSet4 | type | 2654 |  |
| LPP-PDU-Definitions | UTC-ModelSet5-r12 | type | 2667 |  |
| LPP-PDU-Definitions | GNSS-AuxiliaryInformation | type | 2677 | GNSS-ID-GPS, GNSS-ID-GLONASS, GNSS-ID-BDS-r16 |
| LPP-PDU-Definitions | GNSS-ID-GPS | type | 2684 | GNSS-ID-GPS-SatElement |
| LPP-PDU-Definitions | GNSS-ID-GPS-SatElement | type | 2685 | SV-ID, GNSS-SignalIDs |
| LPP-PDU-Definitions | GNSS-ID-GLONASS | type | 2690 | GNSS-ID-GLONASS-SatElement |
| LPP-PDU-Definitions | GNSS-ID-GLONASS-SatElement | type | 2691 | SV-ID, GNSS-SignalIDs |
| LPP-PDU-Definitions | GNSS-ID-BDS-r16 | type | 2697 | GNSS-ID-BDS-SatElement-r16 |
| LPP-PDU-Definitions | GNSS-ID-BDS-SatElement-r16 | type | 2698 | SV-ID |
| LPP-PDU-Definitions | BDS-DifferentialCorrections-r12 | type | 2704 | BDS-SgnTypeList-r12 |
| LPP-PDU-Definitions | BDS-SgnTypeList-r12 | type | 2709 | BDS-SgnTypeElement-r12 |
| LPP-PDU-Definitions | BDS-SgnTypeElement-r12 | type | 2710 | GNSS-SignalID, DBDS-CorrectionList-r12 |
| LPP-PDU-Definitions | DBDS-CorrectionList-r12 | type | 2715 | DBDS-CorrectionElement-r12 |
| LPP-PDU-Definitions | DBDS-CorrectionElement-r12 | type | 2716 | SV-ID |
| LPP-PDU-Definitions | BDS-GridModelParameter-r12 | type | 2724 | GridIonList-r12 |
| LPP-PDU-Definitions | GridIonList-r12 | type | 2729 | GridIonElement-r12 |
| LPP-PDU-Definitions | GridIonElement-r12 | type | 2730 |  |
| LPP-PDU-Definitions | GNSS-RTK-Observations-r15 | type | 2737 | GNSS-SystemTime, GNSS-ObservationList-r15 |
| LPP-PDU-Definitions | GNSS-ObservationList-r15 | type | 2743 | GNSS-RTK-SatelliteDataElement-r15 |
| LPP-PDU-Definitions | GNSS-RTK-SatelliteDataElement-r15 | type | 2744 | SV-ID, GNSS-RTK-SatelliteSignalDataList-r15 |
| LPP-PDU-Definitions | GNSS-RTK-SatelliteSignalDataList-r15 | type | 2752 | GNSS-RTK-SatelliteSignalDataElement-r15 |
| LPP-PDU-Definitions | GNSS-RTK-SatelliteSignalDataElement-r15 | type | 2754 | GNSS-SignalID |
| LPP-PDU-Definitions | GLO-RTK-BiasInformation-r15 | type | 2765 | GNSS-ReferenceStationID-r15 |
| LPP-PDU-Definitions | GNSS-RTK-MAC-CorrectionDifferences-r15 | type | 2775 | GNSS-NetworkID-r15, GNSS-SubNetworkID-r15, GNSS-ReferenceStationID-r15, GNSS-FrequencyID-r15, RTK-CorrectionDifferencesList-r15 |
| LPP-PDU-Definitions | RTK-CorrectionDifferencesList-r15 | type | 2784 | RTK-CorrectionDifferencesElement-r15 |
| LPP-PDU-Definitions | RTK-CorrectionDifferencesElement-r15 | type | 2786 | GNSS-SystemTime, GNSS-ReferenceStationID-r15, Geometric-Ionospheric-Corrections-Differences-r15 |
| LPP-PDU-Definitions | Geometric-Ionospheric-Corrections-Differences-r15 | type | 2793 | Geometric-Ionospheric-Corrections-Differences-Element-r15 |
| LPP-PDU-Definitions | Geometric-Ionospheric-Corrections-Differences-Element-r15 | type | 2795 | SV-ID |
| LPP-PDU-Definitions | GNSS-RTK-Residuals-r15 | type | 2805 | GNSS-SystemTime, GNSS-ReferenceStationID-r15, GNSS-FrequencyID-r15, RTK-Residuals-List-r15 |
| LPP-PDU-Definitions | RTK-Residuals-List-r15 | type | 2814 | RTK-Residuals-Element-r15 |
| LPP-PDU-Definitions | RTK-Residuals-Element-r15 | type | 2815 | SV-ID |
| LPP-PDU-Definitions | GNSS-RTK-FKP-Gradients-r15 | type | 2825 | GNSS-ReferenceStationID-r15, GNSS-SystemTime, GNSS-FrequencyID-r15, FKP-Gradients-List-r15 |
| LPP-PDU-Definitions | FKP-Gradients-List-r15 | type | 2833 | FKP-Gradients-Element-r15 |
| LPP-PDU-Definitions | FKP-Gradients-Element-r15 | type | 2834 | SV-ID |
| LPP-PDU-Definitions | GNSS-SSR-OrbitCorrections-r15 | type | 2844 | GNSS-SystemTime, SSR-OrbitCorrectionList-r15 |
| LPP-PDU-Definitions | SSR-OrbitCorrectionList-r15 | type | 2852 | SSR-OrbitCorrectionSatelliteElement-r15 |
| LPP-PDU-Definitions | SSR-OrbitCorrectionSatelliteElement-r15 | type | 2853 | SV-ID |
| LPP-PDU-Definitions | GNSS-SSR-ClockCorrections-r15 | type | 2865 | GNSS-SystemTime, SSR-ClockCorrectionList-r15 |
| LPP-PDU-Definitions | SSR-ClockCorrectionList-r15 | type | 2872 | SSR-ClockCorrectionSatelliteElement-r15 |
| LPP-PDU-Definitions | SSR-ClockCorrectionSatelliteElement-r15 | type | 2873 | SV-ID |
| LPP-PDU-Definitions | GNSS-SSR-CodeBias-r15 | type | 2881 | GNSS-SystemTime, SSR-CodeBiasSatList-r15 |
| LPP-PDU-Definitions | SSR-CodeBiasSatList-r15 | type | 2888 | SSR-CodeBiasSatElement-r15 |
| LPP-PDU-Definitions | SSR-CodeBiasSatElement-r15 | type | 2889 | SV-ID, SSR-CodeBiasSignalList-r15 |
| LPP-PDU-Definitions | SSR-CodeBiasSignalList-r15 | type | 2894 | SSR-CodeBiasSignalElement-r15 |
| LPP-PDU-Definitions | SSR-CodeBiasSignalElement-r15 | type | 2895 | GNSS-SignalID |
| LPP-PDU-Definitions | GNSS-SSR-URA-r16 | type | 2901 | GNSS-SystemTime, SSR-URA-SatList-r16 |
| LPP-PDU-Definitions | SSR-URA-SatList-r16 | type | 2908 | SSR-URA-SatElement-r16 |
| LPP-PDU-Definitions | SSR-URA-SatElement-r16 | type | 2909 | SV-ID |
| LPP-PDU-Definitions | GNSS-SSR-PhaseBias-r16 | type | 2915 | GNSS-SystemTime, SSR-PhaseBiasSatList-r16 |
| LPP-PDU-Definitions | SSR-PhaseBiasSatList-r16 | type | 2922 | SSR-PhaseBiasSatElement-r16 |
| LPP-PDU-Definitions | SSR-PhaseBiasSatElement-r16 | type | 2923 | SV-ID, SSR-PhaseBiasSignalList-r16 |
| LPP-PDU-Definitions | SSR-PhaseBiasSignalList-r16 | type | 2928 | SSR-PhaseBiasSignalElement-r16 |
| LPP-PDU-Definitions | SSR-PhaseBiasSignalElement-r16 | type | 2929 | GNSS-SignalID |
| LPP-PDU-Definitions | GNSS-SSR-STEC-Correction-r16 | type | 2937 | GNSS-SystemTime, STEC-SatList-r16 |
| LPP-PDU-Definitions | STEC-SatList-r16 | type | 2945 | STEC-SatElement-r16 |
| LPP-PDU-Definitions | STEC-SatElement-r16 | type | 2946 | SV-ID |
| LPP-PDU-Definitions | GNSS-SSR-GriddedCorrection-r16 | type | 2956 | GNSS-SystemTime, GridList-r16 |
| LPP-PDU-Definitions | GridList-r16 | type | 2965 | GridElement-r16 |
| LPP-PDU-Definitions | GridElement-r16 | type | 2966 | TropospericDelayCorrection-r16, STEC-ResidualSatList-r16 |
| LPP-PDU-Definitions | TropospericDelayCorrection-r16 | type | 2971 |  |
| LPP-PDU-Definitions | STEC-ResidualSatList-r16 | type | 2976 | STEC-ResidualSatElement-r16 |
| LPP-PDU-Definitions | STEC-ResidualSatElement-r16 | type | 2977 | SV-ID |
| LPP-PDU-Definitions | NavIC-DifferentialCorrections-r16 | type | 2986 | NavIC-CorrectionListAutoNav-r16 |
| LPP-PDU-Definitions | NavIC-CorrectionListAutoNav-r16 | type | 2991 | NavIC-CorrectionElementAutoNav-r16 |
| LPP-PDU-Definitions | NavIC-CorrectionElementAutoNav-r16 | type | 2992 | SV-ID, NavIC-EDC-r16, NavIC-CDC-r16 |
| LPP-PDU-Definitions | NavIC-EDC-r16 | type | 3002 |  |
| LPP-PDU-Definitions | NavIC-CDC-r16 | type | 3011 |  |
| LPP-PDU-Definitions | NavIC-GridModelParameter-r16 | type | 3018 | RegionIgpList-r16 |
| LPP-PDU-Definitions | RegionIgpList-r16 | type | 3025 | RegionIgpElement-r16 |
| LPP-PDU-Definitions | RegionIgpElement-r16 | type | 3027 |  |
| LPP-PDU-Definitions | A-GNSS-RequestAssistanceData | type | 3062 | GNSS-CommonAssistDataReq, GNSS-GenericAssistDataReq, GNSS-PeriodicAssistDataReq-r15 |
| LPP-PDU-Definitions | GNSS-CommonAssistDataReq | type | 3072 | GNSS-ReferenceTimeReq, GNSS-ReferenceLocationReq, GNSS-IonosphericModelReq, GNSS-EarthOrientationParametersReq, GNSS-RTK-ReferenceStationInfoReq-r15, GNSS-RTK-AuxiliaryStationDataReq-r15, GNSS-SSR-CorrectionPointsReq-r16 |
| LPP-PDU-Definitions | GNSS-GenericAssistDataReq | type | 3097 | GNSS-GenericAssistDataReqElement |
| LPP-PDU-Definitions | GNSS-GenericAssistDataReqElement | type | 3098 | GNSS-ID, SBAS-ID, GNSS-TimeModelListReq, GNSS-DifferentialCorrectionsReq, GNSS-NavigationModelReq, GNSS-RealTimeIntegrityReq, GNSS-DataBitAssistanceReq, GNSS-AcquisitionAssistanceReq, GNSS-AlmanacReq, GNSS-UTC-ModelReq, GNSS-AuxiliaryInformationReq, BDS-DifferentialCorrectionsReq-r12, BDS-GridModelReq-r12, GNSS-RTK-ObservationsReq-r15, GLO-RTK-BiasInformationReq-r15, GNSS-RTK-MAC-CorrectionDifferencesReq-r15, GNSS-RTK-ResidualsReq-r15, GNSS-RTK-FKP-GradientsReq-r15, GNSS-SSR-OrbitCorrectionsReq-r15, GNSS-SSR-ClockCorrectionsReq-r15, GNSS-SSR-CodeBiasReq-r15, GNSS-SSR-URA-Req-r16, GNSS-SSR-PhaseBiasReq-r16, GNSS-SSR-STEC-CorrectionReq-r16, GNSS-SSR-GriddedCorrectionReq-r16, NavIC-DifferentialCorrectionsReq-r16, NavIC-GridModelReq-r16 |
| LPP-PDU-Definitions | GNSS-PeriodicAssistDataReq-r15 | type | 3150 | GNSS-PeriodicControlParam-r15 |
| LPP-PDU-Definitions | GNSS-ReferenceTimeReq | type | 3172 | GNSS-ID |
| LPP-PDU-Definitions | GNSS-ReferenceLocationReq | type | 3179 |  |
| LPP-PDU-Definitions | GNSS-IonosphericModelReq | type | 3183 |  |
| LPP-PDU-Definitions | GNSS-EarthOrientationParametersReq | type | 3191 |  |
| LPP-PDU-Definitions | GNSS-RTK-ReferenceStationInfoReq-r15 | type | 3195 | GNSS-ReferenceStationID-r15 |
| LPP-PDU-Definitions | GNSS-RTK-AuxiliaryStationDataReq-r15 | type | 3203 | GNSS-ReferenceStationID-r15 |
| LPP-PDU-Definitions | GNSS-SSR-CorrectionPointsReq-r16 | type | 3208 |  |
| LPP-PDU-Definitions | GNSS-TimeModelListReq | type | 3213 | GNSS-TimeModelElementReq |
| LPP-PDU-Definitions | GNSS-TimeModelElementReq | type | 3214 |  |
| LPP-PDU-Definitions | GNSS-DifferentialCorrectionsReq | type | 3220 | GNSS-SignalIDs |
| LPP-PDU-Definitions | GNSS-NavigationModelReq | type | 3226 | StoredNavListInfo, ReqNavListInfo |
| LPP-PDU-Definitions | StoredNavListInfo | type | 3231 | SatListRelatedDataList |
| LPP-PDU-Definitions | SatListRelatedDataList | type | 3238 | SatListRelatedDataElement |
| LPP-PDU-Definitions | SatListRelatedDataElement | type | 3239 | SV-ID |
| LPP-PDU-Definitions | ReqNavListInfo | type | 3246 |  |
| LPP-PDU-Definitions | GNSS-RealTimeIntegrityReq | type | 3254 |  |
| LPP-PDU-Definitions | GNSS-DataBitAssistanceReq | type | 3258 | GNSS-SignalIDs, GNSS-DataBitsReqSatList |
| LPP-PDU-Definitions | GNSS-DataBitsReqSatList | type | 3266 | GNSS-DataBitsReqSatElement |
| LPP-PDU-Definitions | GNSS-DataBitsReqSatElement | type | 3267 | SV-ID |
| LPP-PDU-Definitions | GNSS-AcquisitionAssistanceReq | type | 3272 | GNSS-SignalID |
| LPP-PDU-Definitions | GNSS-AlmanacReq | type | 3277 |  |
| LPP-PDU-Definitions | GNSS-UTC-ModelReq | type | 3282 |  |
| LPP-PDU-Definitions | GNSS-AuxiliaryInformationReq | type | 3287 |  |
| LPP-PDU-Definitions | BDS-DifferentialCorrectionsReq-r12 | type | 3291 | GNSS-SignalIDs |
| LPP-PDU-Definitions | BDS-GridModelReq-r12 | type | 3296 |  |
| LPP-PDU-Definitions | GNSS-RTK-ObservationsReq-r15 | type | 3300 | GNSS-SignalIDs, GNSS-ReferenceStationID-r15 |
| LPP-PDU-Definitions | GLO-RTK-BiasInformationReq-r15 | type | 3309 | GNSS-ReferenceStationID-r15 |
| LPP-PDU-Definitions | GNSS-RTK-MAC-CorrectionDifferencesReq-r15 | type | 3314 | GNSS-ReferenceStationID-r15, AUX-ReferenceStationList-r15, GNSS-Link-CombinationsList-r15 |
| LPP-PDU-Definitions | AUX-ReferenceStationList-r15 | type | 3320 | AUX-ReferenceStationID-Element-r15 |
| LPP-PDU-Definitions | AUX-ReferenceStationID-Element-r15 | type | 3321 | GNSS-ReferenceStationID-r15 |
| LPP-PDU-Definitions | GNSS-RTK-ResidualsReq-r15 | type | 3326 | GNSS-ReferenceStationID-r15, GNSS-Link-CombinationsList-r15 |
| LPP-PDU-Definitions | GNSS-RTK-FKP-GradientsReq-r15 | type | 3332 | GNSS-ReferenceStationID-r15, GNSS-Link-CombinationsList-r15 |
| LPP-PDU-Definitions | GNSS-SSR-OrbitCorrectionsReq-r15 | type | 3338 | GNSS-NavListInfo-r15 |
| LPP-PDU-Definitions | GNSS-SSR-ClockCorrectionsReq-r15 | type | 3343 | GNSS-NavListInfo-r15 |
| LPP-PDU-Definitions | GNSS-SSR-CodeBiasReq-r15 | type | 3348 | GNSS-SignalIDs, GNSS-NavListInfo-r15 |
| LPP-PDU-Definitions | GNSS-SSR-URA-Req-r16 | type | 3354 |  |
| LPP-PDU-Definitions | GNSS-SSR-PhaseBiasReq-r16 | type | 3358 | GNSS-SignalIDs, GNSS-NavListInfo-r15 |
| LPP-PDU-Definitions | GNSS-SSR-STEC-CorrectionReq-r16 | type | 3364 |  |
| LPP-PDU-Definitions | GNSS-SSR-GriddedCorrectionReq-r16 | type | 3368 |  |
| LPP-PDU-Definitions | NavIC-DifferentialCorrectionsReq-r16 | type | 3372 | GNSS-SignalIDs |
| LPP-PDU-Definitions | NavIC-GridModelReq-r16 | type | 3377 |  |
| LPP-PDU-Definitions | A-GNSS-ProvideLocationInformation | type | 3381 | GNSS-SignalMeasurementInformation, GNSS-LocationInformation, A-GNSS-Error |
| LPP-PDU-Definitions | GNSS-SignalMeasurementInformation | type | 3388 | MeasurementReferenceTime, GNSS-MeasurementList |
| LPP-PDU-Definitions | MeasurementReferenceTime | type | 3394 | GNSS-ID, CellGlobalIdEUTRA-AndUTRA, CellGlobalIdGERAN, ECGI, NCGI-r15 |
| LPP-PDU-Definitions | GNSS-MeasurementList | type | 3453 | GNSS-MeasurementForOneGNSS |
| LPP-PDU-Definitions | GNSS-MeasurementForOneGNSS | type | 3454 | GNSS-ID, GNSS-SgnMeasList |
| LPP-PDU-Definitions | GNSS-SgnMeasList | type | 3459 | GNSS-SgnMeasElement |
| LPP-PDU-Definitions | GNSS-SgnMeasElement | type | 3460 | GNSS-SignalID, GNSS-SatMeasList |
| LPP-PDU-Definitions | GNSS-SatMeasList | type | 3466 | GNSS-SatMeasElement |
| LPP-PDU-Definitions | GNSS-SatMeasElement | type | 3467 | SV-ID |
| LPP-PDU-Definitions | GNSS-LocationInformation | type | 3486 | MeasurementReferenceTime, GNSS-ID-Bitmap |
| LPP-PDU-Definitions | A-GNSS-RequestLocationInformation | type | 3492 | GNSS-PositioningInstructions |
| LPP-PDU-Definitions | GNSS-PositioningInstructions | type | 3497 | GNSS-ID-Bitmap |
| LPP-PDU-Definitions | A-GNSS-ProvideCapabilities | type | 3509 | GNSS-SupportList, AssistanceDataSupportList, LocationCoordinateTypes, VelocityTypes, PositioningModes |
| LPP-PDU-Definitions | GNSS-SupportList | type | 3525 | GNSS-SupportElement |
| LPP-PDU-Definitions | GNSS-SupportElement | type | 3526 | GNSS-ID, SBAS-IDs, PositioningModes, GNSS-SignalIDs, AccessTypes |
| LPP-PDU-Definitions | AssistanceDataSupportList | type | 3544 | GNSS-CommonAssistanceDataSupport, GNSS-GenericAssistanceDataSupport |
| LPP-PDU-Definitions | GNSS-CommonAssistanceDataSupport | type | 3550 | GNSS-ReferenceTimeSupport, GNSS-ReferenceLocationSupport, GNSS-IonosphericModelSupport, GNSS-EarthOrientationParametersSupport, GNSS-RTK-ReferenceStationInfoSupport-r15, GNSS-RTK-AuxiliaryStationDataSupport-r15 |
| LPP-PDU-Definitions | GNSS-ReferenceTimeSupport | type | 3570 | GNSS-ID-Bitmap, AccessTypes |
| LPP-PDU-Definitions | GNSS-ReferenceLocationSupport | type | 3576 |  |
| LPP-PDU-Definitions | GNSS-IonosphericModelSupport | type | 3580 |  |
| LPP-PDU-Definitions | GNSS-EarthOrientationParametersSupport | type | 3587 |  |
| LPP-PDU-Definitions | GNSS-RTK-ReferenceStationInfoSupport-r15 | type | 3591 |  |
| LPP-PDU-Definitions | GNSS-RTK-AuxiliaryStationDataSupport-r15 | type | 3595 |  |
| LPP-PDU-Definitions | GNSS-GenericAssistanceDataSupport | type | 3599 | GNSS-GenericAssistDataSupportElement |
| LPP-PDU-Definitions | GNSS-GenericAssistDataSupportElement | type | 3601 | GNSS-ID, SBAS-ID, GNSS-TimeModelListSupport, GNSS-DifferentialCorrectionsSupport, GNSS-NavigationModelSupport, GNSS-RealTimeIntegritySupport, GNSS-DataBitAssistanceSupport, GNSS-AcquisitionAssistanceSupport, GNSS-AlmanacSupport, GNSS-UTC-ModelSupport, GNSS-AuxiliaryInformationSupport, BDS-DifferentialCorrectionsSupport-r12, BDS-GridModelSupport-r12, GNSS-RTK-ObservationsSupport-r15, GLO-RTK-BiasInformationSupport-r15, GNSS-RTK-MAC-CorrectionDifferencesSupport-r15, GNSS-RTK-ResidualsSupport-r15, GNSS-RTK-FKP-GradientsSupport-r15, GNSS-SSR-OrbitCorrectionsSupport-r15, GNSS-SSR-ClockCorrectionsSupport-r15, GNSS-SSR-CodeBiasSupport-r15, GNSS-SSR-URA-Support-r16, GNSS-SSR-PhaseBiasSupport-r16, GNSS-SSR-STEC-CorrectionSupport-r16, GNSS-SSR-GriddedCorrectionSupport-r16, NavIC-DifferentialCorrectionsSupport-r16, NavIC-GridModelSupport-r16 |
| LPP-PDU-Definitions | GNSS-TimeModelListSupport | type | 3672 |  |
| LPP-PDU-Definitions | GNSS-DifferentialCorrectionsSupport | type | 3676 | GNSS-SignalIDs |
| LPP-PDU-Definitions | GNSS-NavigationModelSupport | type | 3682 |  |
| LPP-PDU-Definitions | GNSS-RealTimeIntegritySupport | type | 3702 |  |
| LPP-PDU-Definitions | GNSS-DataBitAssistanceSupport | type | 3706 |  |
| LPP-PDU-Definitions | GNSS-AcquisitionAssistanceSupport | type | 3710 |  |
| LPP-PDU-Definitions | GNSS-AlmanacSupport | type | 3716 |  |
| LPP-PDU-Definitions | GNSS-UTC-ModelSupport | type | 3727 |  |
| LPP-PDU-Definitions | GNSS-AuxiliaryInformationSupport | type | 3736 |  |
| LPP-PDU-Definitions | BDS-DifferentialCorrectionsSupport-r12 | type | 3740 | GNSS-SignalIDs |
| LPP-PDU-Definitions | BDS-GridModelSupport-r12 | type | 3745 |  |
| LPP-PDU-Definitions | GNSS-RTK-ObservationsSupport-r15 | type | 3749 | GNSS-SignalIDs |
| LPP-PDU-Definitions | GLO-RTK-BiasInformationSupport-r15 | type | 3754 |  |
| LPP-PDU-Definitions | GNSS-RTK-MAC-CorrectionDifferencesSupport-r15 | type | 3758 | GNSS-Link-CombinationsList-r15 |
| LPP-PDU-Definitions | GNSS-RTK-ResidualsSupport-r15 | type | 3763 | GNSS-Link-CombinationsList-r15 |
| LPP-PDU-Definitions | GNSS-RTK-FKP-GradientsSupport-r15 | type | 3768 | GNSS-Link-CombinationsList-r15 |
| LPP-PDU-Definitions | GNSS-SSR-OrbitCorrectionsSupport-r15 | type | 3773 |  |
| LPP-PDU-Definitions | GNSS-SSR-ClockCorrectionsSupport-r15 | type | 3777 |  |
| LPP-PDU-Definitions | GNSS-SSR-CodeBiasSupport-r15 | type | 3781 | GNSS-SignalIDs |
| LPP-PDU-Definitions | GNSS-SSR-URA-Support-r16 | type | 3786 |  |
| LPP-PDU-Definitions | GNSS-SSR-PhaseBiasSupport-r16 | type | 3790 | GNSS-SignalIDs |
| LPP-PDU-Definitions | GNSS-SSR-STEC-CorrectionSupport-r16 | type | 3795 |  |
| LPP-PDU-Definitions | GNSS-SSR-GriddedCorrectionSupport-r16 | type | 3799 |  |
| LPP-PDU-Definitions | NavIC-DifferentialCorrectionsSupport-r16 | type | 3803 | GNSS-SignalIDs |
| LPP-PDU-Definitions | NavIC-GridModelSupport-r16 | type | 3808 |  |
| LPP-PDU-Definitions | A-GNSS-RequestCapabilities | type | 3812 |  |
| LPP-PDU-Definitions | A-GNSS-Error | type | 3819 | GNSS-LocationServerErrorCauses, GNSS-TargetDeviceErrorCauses |
| LPP-PDU-Definitions | GNSS-LocationServerErrorCauses | type | 3825 |  |
| LPP-PDU-Definitions | GNSS-TargetDeviceErrorCauses | type | 3839 |  |
| LPP-PDU-Definitions | GNSS-FrequencyID-r15 | type | 3852 |  |
| LPP-PDU-Definitions | GNSS-ID | type | 3857 |  |
| LPP-PDU-Definitions | GNSS-ID-Bitmap | type | 3862 |  |
| LPP-PDU-Definitions | GNSS-Link-CombinationsList-r15 | type | 3873 | GNSS-Link-Combinations-r15 |
| LPP-PDU-Definitions | GNSS-Link-Combinations-r15 | type | 3874 | GNSS-FrequencyID-r15 |
| LPP-PDU-Definitions | GNSS-NavListInfo-r15 | type | 3880 | SatListElement-r15 |
| LPP-PDU-Definitions | SatListElement-r15 | type | 3881 | SV-ID |
| LPP-PDU-Definitions | GNSS-NetworkID-r15 | type | 3887 |  |
| LPP-PDU-Definitions | GNSS-PeriodicControlParam-r15 | type | 3892 |  |
| LPP-PDU-Definitions | GNSS-ReferenceStationID-r15 | type | 3898 |  |
| LPP-PDU-Definitions | GNSS-SignalID | type | 3904 |  |
| LPP-PDU-Definitions | GNSS-SignalIDs | type | 3912 |  |
| LPP-PDU-Definitions | GNSS-SubNetworkID-r15 | type | 3920 |  |
| LPP-PDU-Definitions | SBAS-ID | type | 3925 |  |
| LPP-PDU-Definitions | SBAS-IDs | type | 3930 |  |
| LPP-PDU-Definitions | SV-ID | type | 3938 |  |
| LPP-PDU-Definitions | ECID-ProvideLocationInformation | type | 3943 | ECID-SignalMeasurementInformation, ECID-Error |
| LPP-PDU-Definitions | ECID-SignalMeasurementInformation | type | 3949 | MeasuredResultsElement, MeasuredResultsList |
| LPP-PDU-Definitions | MeasuredResultsList | type | 3954 | MeasuredResultsElement |
| LPP-PDU-Definitions | MeasuredResultsElement | type | 3955 | CellGlobalIdEUTRA-AndUTRA, ARFCN-ValueEUTRA, ARFCN-ValueEUTRA-v9a0, CarrierFreqOffsetNB-r14 |
| LPP-PDU-Definitions | ECID-RequestLocationInformation | type | 3977 |  |
| LPP-PDU-Definitions | ECID-ProvideCapabilities | type | 3986 |  |
| LPP-PDU-Definitions | ECID-RequestCapabilities | type | 4001 |  |
| LPP-PDU-Definitions | ECID-Error | type | 4005 | ECID-LocationServerErrorCauses, ECID-TargetDeviceErrorCauses |
| LPP-PDU-Definitions | ECID-LocationServerErrorCauses | type | 4011 |  |
| LPP-PDU-Definitions | ECID-TargetDeviceErrorCauses | type | 4018 |  |
| LPP-PDU-Definitions | TBS-ProvideLocationInformation-r13 | type | 4034 | TBS-MeasurementInformation-r13, TBS-Error-r13 |
| LPP-PDU-Definitions | TBS-MeasurementInformation-r13 | type | 4040 | MBS-BeaconMeasList-r13 |
| LPP-PDU-Definitions | MBS-BeaconMeasList-r13 | type | 4046 | MBS-BeaconMeasElement-r13 |
| LPP-PDU-Definitions | MBS-BeaconMeasElement-r13 | type | 4047 |  |
| LPP-PDU-Definitions | TBS-RequestLocationInformation-r13 | type | 4056 |  |
| LPP-PDU-Definitions | TBS-ProvideCapabilities-r13 | type | 4065 | MBS-AssistanceDataSupportList-r14, PositioningModes |
| LPP-PDU-Definitions | MBS-AssistanceDataSupportList-r14 | type | 4080 |  |
| LPP-PDU-Definitions | TBS-RequestCapabilities-r13 | type | 4086 |  |
| LPP-PDU-Definitions | TBS-Error-r13 | type | 4090 | TBS-LocationServerErrorCauses-r13, TBS-TargetDeviceErrorCauses-r13 |
| LPP-PDU-Definitions | TBS-LocationServerErrorCauses-r13 | type | 4096 |  |
| LPP-PDU-Definitions | TBS-TargetDeviceErrorCauses-r13 | type | 4105 |  |
| LPP-PDU-Definitions | TBS-ProvideAssistanceData-r14 | type | 4114 | TBS-AssistanceDataList-r14, TBS-Error-r13 |
| LPP-PDU-Definitions | TBS-AssistanceDataList-r14 | type | 4120 | MBS-AssistanceDataList-r14 |
| LPP-PDU-Definitions | MBS-AssistanceDataList-r14 | type | 4124 | maxMBS-r14, MBS-AssistanceDataElement-r14 |
| LPP-PDU-Definitions | MBS-AssistanceDataElement-r14 | type | 4125 | MBS-AlmanacAssistance-r14, MBS-AcquisitionAssistance-r14 |
| LPP-PDU-Definitions | MBS-AlmanacAssistance-r14 | type | 4131 |  |
| LPP-PDU-Definitions | MBS-AcquisitionAssistance-r14 | type | 4140 |  |
| LPP-PDU-Definitions | TBS-RequestAssistanceData-r14 | type | 4148 |  |
| LPP-PDU-Definitions | Sensor-ProvideLocationInformation-r13 | type | 4154 | Sensor-MeasurementInformation-r13, Sensor-Error-r13, Sensor-MotionInformation-r15 |
| LPP-PDU-Definitions | Sensor-MeasurementInformation-r13 | type | 4163 |  |
| LPP-PDU-Definitions | Sensor-MotionInformation-r15 | type | 4177 | DisplacementTimeStamp-r15, DisplacementInfoList-r15 |
| LPP-PDU-Definitions | DisplacementInfoList-r15 | type | 4182 | DisplacementInfoListElement-r15 |
| LPP-PDU-Definitions | DisplacementInfoListElement-r15 | type | 4183 | DeltaTime-r15, Displacement-r15 |
| LPP-PDU-Definitions | DisplacementTimeStamp-r15 | type | 4188 | UTC-Time-r15, MeasurementReferenceTime, SFN-r15 |
| LPP-PDU-Definitions | DeltaTime-r15 | type | 4195 |  |
| LPP-PDU-Definitions | SFN-r15 | type | 4200 |  |
| LPP-PDU-Definitions | Displacement-r15 | type | 4205 |  |
| LPP-PDU-Definitions | UTC-Time-r15 | type | 4218 |  |
| LPP-PDU-Definitions | Sensor-RequestLocationInformation-r13 | type | 4224 |  |
| LPP-PDU-Definitions | Sensor-ProvideCapabilities-r13 | type | 4235 | Sensor-AssistanceDataSupportList-r14, PositioningModes |
| LPP-PDU-Definitions | Sensor-AssistanceDataSupportList-r14 | type | 4249 |  |
| LPP-PDU-Definitions | Sensor-RequestCapabilities-r13 | type | 4256 |  |
| LPP-PDU-Definitions | Sensor-Error-r13 | type | 4260 | Sensor-LocationServerErrorCauses-r13, Sensor-TargetDeviceErrorCauses-r13 |
| LPP-PDU-Definitions | Sensor-LocationServerErrorCauses-r13 | type | 4266 |  |
| LPP-PDU-Definitions | Sensor-TargetDeviceErrorCauses-r13 | type | 4275 |  |
| LPP-PDU-Definitions | Sensor-ProvideAssistanceData-r14 | type | 4283 | Sensor-AssistanceDataList-r14, Sensor-Error-r13 |
| LPP-PDU-Definitions | Sensor-AssistanceDataList-r14 | type | 4289 | EllipsoidPointWithAltitudeAndUncertaintyEllipsoid, PressureValidityPeriod-v1520, PressureValidityArea-v1520 |
| LPP-PDU-Definitions | PressureValidityArea-v1520 | type | 4308 | Ellipsoid-Point |
| LPP-PDU-Definitions | PressureValidityPeriod-v1520 | type | 4314 | GNSS-SystemTime |
| LPP-PDU-Definitions | Sensor-RequestAssistanceData-r14 | type | 4321 |  |
| LPP-PDU-Definitions | WLAN-ProvideLocationInformation-r13 | type | 4325 | WLAN-MeasurementInformation-r13, WLAN-Error-r13 |
| LPP-PDU-Definitions | WLAN-MeasurementInformation-r13 | type | 4331 | WLAN-MeasurementList-r13 |
| LPP-PDU-Definitions | WLAN-MeasurementList-r13 | type | 4336 | maxWLAN-AP-r13, WLAN-MeasurementElement-r13 |
| LPP-PDU-Definitions | WLAN-MeasurementElement-r13 | type | 4337 | WLAN-AP-Identifier-r13, WLAN-RTT-r13 |
| LPP-PDU-Definitions | WLAN-AP-Identifier-r13 | type | 4345 |  |
| LPP-PDU-Definitions | WLAN-RTT-r13 | type | 4350 |  |
| LPP-PDU-Definitions | WLAN-RequestLocationInformation-r13 | type | 4362 |  |
| LPP-PDU-Definitions | WLAN-ProvideCapabilities-r13 | type | 4371 | PositioningModes |
| LPP-PDU-Definitions | WLAN-RequestCapabilities-r13 | type | 4389 |  |
| LPP-PDU-Definitions | WLAN-Error-r13 | type | 4393 | WLAN-LocationServerErrorCauses-r13, WLAN-TargetDeviceErrorCauses-r13 |
| LPP-PDU-Definitions | WLAN-LocationServerErrorCauses-r13 | type | 4399 |  |
| LPP-PDU-Definitions | WLAN-TargetDeviceErrorCauses-r13 | type | 4410 |  |
| LPP-PDU-Definitions | WLAN-ProvideAssistanceData-r14 | type | 4421 | maxWLAN-DataSets-r14, WLAN-DataSet-r14, WLAN-Error-r13 |
| LPP-PDU-Definitions | WLAN-DataSet-r14 | type | 4428 | maxWLAN-AP-r14, WLAN-AP-Data-r14, SupportedChannels-11a-r14, SupportedChannels-11bg-r14 |
| LPP-PDU-Definitions | SupportedChannels-11a-r14 | type | 4434 |  |
| LPP-PDU-Definitions | SupportedChannels-11bg-r14 | type | 4452 |  |
| LPP-PDU-Definitions | WLAN-AP-Data-r14 | type | 4469 | WLAN-AP-Identifier-r13, WLAN-AP-Location-r14 |
| LPP-PDU-Definitions | WLAN-AP-Location-r14 | type | 4474 | LocationDataLCI-r14 |
| LPP-PDU-Definitions | LocationDataLCI-r14 | type | 4478 |  |
| LPP-PDU-Definitions | WLAN-RequestAssistanceData-r14 | type | 4489 | maxVisibleAPs-r14, WLAN-AP-Identifier-r13, maxKnownAPs-r14 |
| LPP-PDU-Definitions | BT-ProvideLocationInformation-r13 | type | 4497 | BT-MeasurementInformation-r13, BT-Error-r13 |
| LPP-PDU-Definitions | BT-MeasurementInformation-r13 | type | 4503 | BT-MeasurementList-r13 |
| LPP-PDU-Definitions | BT-MeasurementList-r13 | type | 4508 | maxBT-Beacon-r13, BT-MeasurementElement-r13 |
| LPP-PDU-Definitions | BT-MeasurementElement-r13 | type | 4509 |  |
| LPP-PDU-Definitions | BT-RequestLocationInformation-r13 | type | 4515 |  |
| LPP-PDU-Definitions | BT-ProvideCapabilities-r13 | type | 4521 | PositioningModes |
| LPP-PDU-Definitions | BT-RequestCapabilities-r13 | type | 4534 |  |
| LPP-PDU-Definitions | BT-Error-r13 | type | 4538 | BT-LocationServerErrorCauses-r13, BT-TargetDeviceErrorCauses-r13 |
| LPP-PDU-Definitions | BT-LocationServerErrorCauses-r13 | type | 4544 |  |
| LPP-PDU-Definitions | BT-TargetDeviceErrorCauses-r13 | type | 4549 |  |
| LPP-PDU-Definitions | NR-UL-ProvideCapabilities-r16 | type | 4559 | NR-UL-SRS-Capability-r16 |
| LPP-PDU-Definitions | NR-UL-RequestCapabilities-r16 | type | 4564 |  |
| LPP-PDU-Definitions | NR-ECID-ProvideLocationInformation-r16 | type | 4568 | NR-ECID-SignalMeasurementInformation-r16, NR-ECID-Error-r16 |
| LPP-PDU-Definitions | NR-ECID-SignalMeasurementInformation-r16 | type | 4574 | NR-MeasuredResultsElement-r16, NR-MeasuredResultsList-r16 |
| LPP-PDU-Definitions | NR-MeasuredResultsList-r16 | type | 4579 | NR-MeasuredResultsElement-r16 |
| LPP-PDU-Definitions | NR-MeasuredResultsElement-r16 | type | 4580 | NR-PhysCellID-r16, ARFCN-ValueNR-r15, NCGI-r15, MeasQuantityResults-r16, ResultsPerSSB-IndexList-r16, ResultsPerCSI-RS-IndexList-r16 |
| LPP-PDU-Definitions | MeasQuantityResults-r16 | type | 4593 |  |
| LPP-PDU-Definitions | ResultsPerSSB-IndexList-r16 | type | 4597 | ResultsPerSSB-Index-r16 |
| LPP-PDU-Definitions | ResultsPerSSB-Index-r16 | type | 4598 | MeasQuantityResults-r16 |
| LPP-PDU-Definitions | ResultsPerCSI-RS-IndexList-r16 | type | 4602 | ResultsPerCSI-RS-Index-r16 |
| LPP-PDU-Definitions | ResultsPerCSI-RS-Index-r16 | type | 4603 | MeasQuantityResults-r16 |
| LPP-PDU-Definitions | NR-ECID-RequestLocationInformation-r16 | type | 4608 |  |
| LPP-PDU-Definitions | NR-ECID-ProvideCapabilities-r16 | type | 4616 |  |
| LPP-PDU-Definitions | NR-ECID-RequestCapabilities-r16 | type | 4626 |  |
| LPP-PDU-Definitions | NR-ECID-Error-r16 | type | 4630 | NR-ECID-LocationServerErrorCauses-r16, NR-ECID-TargetDeviceErrorCauses-r16 |
| LPP-PDU-Definitions | NR-ECID-LocationServerErrorCauses-r16 | type | 4636 |  |
| LPP-PDU-Definitions | NR-ECID-TargetDeviceErrorCauses-r16 | type | 4643 |  |
| LPP-PDU-Definitions | NR-DL-TDOA-ProvideAssistanceData-r16 | type | 4656 | NR-DL-PRS-AssistanceData-r16, NR-SelectedDL-PRS-IndexList-r16, NR-PositionCalculationAssistance-r16, NR-DL-TDOA-Error-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-RequestAssistanceData-r16 | type | 4666 | NR-PhysCellID-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-ProvideLocationInformation-r16 | type | 4673 | NR-DL-TDOA-SignalMeasurementInformation-r16, NR-DL-TDOA-LocationInformation-r16, NR-DL-TDOA-Error-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-SignalMeasurementInformation-r16 | type | 4683 | DL-PRS-ID-Info-r16, NR-DL-TDOA-MeasList-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-MeasList-r16 | type | 4688 | nrMaxTRPs-r16, NR-DL-TDOA-MeasElement-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-MeasElement-r16 | type | 4689 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15, NR-DL-PRS-ResourceID-r16, NR-DL-PRS-ResourceSetID-r16, NR-TimeStamp-r16, NR-AdditionalPathList-r16, NR-TimingQuality-r16, NR-DL-TDOA-AdditionalMeasurements-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-AdditionalMeasurements-r16 | type | 4713 | NR-DL-TDOA-AdditionalMeasurementElement-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-AdditionalMeasurementElement-r16 | type | 4715 | NR-DL-PRS-ResourceID-r16, NR-DL-PRS-ResourceSetID-r16, NR-TimeStamp-r16, NR-TimingQuality-r16, NR-AdditionalPathList-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-LocationInformation-r16 | type | 4734 | NR-TimeStamp-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-RequestLocationInformation-r16 | type | 4743 | NR-DL-TDOA-ReportConfig-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-ReportConfig-r16 | type | 4751 |  |
| LPP-PDU-Definitions | NR-DL-TDOA-ProvideCapabilities-r16 | type | 4757 | PositioningModes, NR-DL-PRS-ResourcesCapability-r16, NR-DL-TDOA-MeasurementCapability-r16, NR-DL-PRS-QCL-ProcessingCapability-r16, NR-DL-PRS-ProcessingCapability-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-MeasurementCapability-r16 | type | 4768 |  |
| LPP-PDU-Definitions | NR-DL-TDOA-RequestCapabilities-r16 | type | 4776 |  |
| LPP-PDU-Definitions | NR-DL-TDOA-Error-r16 | type | 4780 | NR-DL-TDOA-LocationServerErrorCauses-r16, NR-DL-TDOA-TargetDeviceErrorCauses-r16 |
| LPP-PDU-Definitions | NR-DL-TDOA-LocationServerErrorCauses-r16 | type | 4786 |  |
| LPP-PDU-Definitions | NR-DL-TDOA-TargetDeviceErrorCauses-r16 | type | 4796 |  |
| LPP-PDU-Definitions | NR-DL-AoD-ProvideAssistanceData-r16 | type | 4808 | NR-DL-PRS-AssistanceData-r16, NR-SelectedDL-PRS-IndexList-r16, NR-PositionCalculationAssistance-r16, NR-DL-AoD-Error-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-RequestAssistanceData-r16 | type | 4818 | NR-PhysCellID-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-ProvideLocationInformation-r16 | type | 4825 | NR-DL-AoD-SignalMeasurementInformation-r16, NR-DL-AoD-LocationInformation-r16, NR-DL-AoD-Error-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-SignalMeasurementInformation-r16 | type | 4835 | NR-DL-AoD-MeasList-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-MeasList-r16 | type | 4839 | nrMaxTRPs-r16, NR-DL-AoD-MeasElement-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-MeasElement-r16 | type | 4840 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15, NR-DL-PRS-ResourceID-r16, NR-DL-PRS-ResourceSetID-r16, NR-TimeStamp-r16, NR-DL-AoD-AdditionalMeasurements-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-AdditionalMeasurements-r16 | type | 4854 | NR-DL-AoD-AdditionalMeasurementElement-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-AdditionalMeasurementElement-r16 | type | 4856 | NR-DL-PRS-ResourceID-r16, NR-DL-PRS-ResourceSetID-r16, NR-TimeStamp-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-LocationInformation-r16 | type | 4865 | NR-TimeStamp-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-RequestLocationInformation-r16 | type | 4874 | NR-DL-AoD-ReportConfig-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-ReportConfig-r16 | type | 4879 |  |
| LPP-PDU-Definitions | NR-DL-AoD-ProvideCapabilities-r16 | type | 4884 | PositioningModes, NR-DL-PRS-ResourcesCapability-r16, NR-DL-AoD-MeasurementCapability-r16, NR-DL-PRS-QCL-ProcessingCapability-r16, NR-DL-PRS-ProcessingCapability-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-MeasurementCapability-r16 | type | 4894 | nrMaxBands-r16, DL-AoD-MeasCapabilityPerBand-r16 |
| LPP-PDU-Definitions | DL-AoD-MeasCapabilityPerBand-r16 | type | 4901 | FreqBandIndicatorNR-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-RequestCapabilities-r16 | type | 4908 |  |
| LPP-PDU-Definitions | NR-DL-AoD-Error-r16 | type | 4912 | NR-DL-AoD-LocationServerErrorCauses-r16, NR-DL-AoD-TargetDeviceErrorCauses-r16 |
| LPP-PDU-Definitions | NR-DL-AoD-LocationServerErrorCauses-r16 | type | 4918 |  |
| LPP-PDU-Definitions | NR-DL-AoD-TargetDeviceErrorCauses-r16 | type | 4928 |  |
| LPP-PDU-Definitions | NR-Multi-RTT-ProvideAssistanceData-r16 | type | 4940 | NR-DL-PRS-AssistanceData-r16, NR-SelectedDL-PRS-IndexList-r16, NR-Multi-RTT-Error-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-RequestAssistanceData-r16 | type | 4947 | NR-PhysCellID-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-ProvideLocationInformation-r16 | type | 4954 | NR-Multi-RTT-SignalMeasurementInformation-r16, NR-Multi-RTT-Error-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-SignalMeasurementInformation-r16 | type | 4962 | NR-Multi-RTT-MeasList-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-MeasList-r16 | type | 4967 | nrMaxTRPs-r16, NR-Multi-RTT-MeasElement-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-MeasElement-r16 | type | 4968 | NR-PhysCellID-r16, NCGI-r15, ARFCN-ValueNR-r15, NR-DL-PRS-ResourceID-r16, NR-DL-PRS-ResourceSetID-r16, NR-AdditionalPathList-r16, NR-TimeStamp-r16, NR-TimingQuality-r16, NR-Multi-RTT-AdditionalMeasurements-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-AdditionalMeasurements-r16 | type | 4992 | NR-Multi-RTT-AdditionalMeasurementElement-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-AdditionalMeasurementElement-r16 | type | 4994 | NR-DL-PRS-ResourceID-r16, NR-DL-PRS-ResourceSetID-r16, NR-TimingQuality-r16, NR-AdditionalPathList-r16, NR-TimeStamp-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-RequestLocationInformation-r16 | type | 5013 | NR-Multi-RTT-ReportConfig-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-ReportConfig-r16 | type | 5022 |  |
| LPP-PDU-Definitions | NR-Multi-RTT-ProvideCapabilities-r16 | type | 5027 | NR-DL-PRS-ResourcesCapability-r16, NR-Multi-RTT-MeasurementCapability-r16, NR-DL-PRS-QCL-ProcessingCapability-r16, NR-DL-PRS-ProcessingCapability-r16, NR-UL-SRS-Capability-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-MeasurementCapability-r16 | type | 5038 |  |
| LPP-PDU-Definitions | NR-Multi-RTT-RequestCapabilities-r16 | type | 5048 |  |
| LPP-PDU-Definitions | NR-Multi-RTT-Error-r16 | type | 5052 | NR-Multi-RTT-LocationServerErrorCauses-r16, NR-Multi-RTT-TargetDeviceErrorCauses-r16 |
| LPP-PDU-Definitions | NR-Multi-RTT-LocationServerErrorCauses-r16 | type | 5058 |  |
| LPP-PDU-Definitions | NR-Multi-RTT-TargetDeviceErrorCauses-r16 | type | 5067 |  |
| LPP-PDU-Definitions | maxEARFCN | value | 5079 |  |
| LPP-PDU-Definitions | maxEARFCN-Plus1 | value | 5080 |  |
| LPP-PDU-Definitions | maxEARFCN2 | value | 5081 |  |
| LPP-PDU-Definitions | maxMBS-r14 | value | 5082 |  |
| LPP-PDU-Definitions | maxWLAN-AP-r13 | value | 5083 |  |
| LPP-PDU-Definitions | maxKnownAPs-r14 | value | 5084 |  |
| LPP-PDU-Definitions | maxVisibleAPs-r14 | value | 5085 |  |
| LPP-PDU-Definitions | maxWLAN-AP-r14 | value | 5086 |  |
| LPP-PDU-Definitions | maxWLAN-DataSets-r14 | value | 5087 |  |
| LPP-PDU-Definitions | maxBT-Beacon-r13 | value | 5088 |  |
| LPP-PDU-Definitions | nrMaxBands-r16 | value | 5089 |  |
| LPP-PDU-Definitions | nrMaxFreqLayers-r16 | value | 5091 |  |
| LPP-PDU-Definitions | nrMaxFreqLayers-1-r16 | value | 5092 |  |
| LPP-PDU-Definitions | nrMaxNumDL-PRS-ResourcesPerSet-1-r16 | value | 5093 |  |
| LPP-PDU-Definitions | nrMaxNumDL-PRS-ResourceSetsPerTRP-1-r16 | value | 5094 |  |
| LPP-PDU-Definitions | nrMaxResourceIDs-r16 | value | 5095 |  |
| LPP-PDU-Definitions | nrMaxResourceOffsetValue-1-r16 | value | 5096 |  |
| LPP-PDU-Definitions | nrMaxResourcesPerSet-r16 | value | 5097 |  |
| LPP-PDU-Definitions | nrMaxSetsPerTrp-r16 | value | 5098 |  |
| LPP-PDU-Definitions | nrMaxSetsPerTrp-1-r16 | value | 5099 |  |
| LPP-PDU-Definitions | nrMaxTRPs-r16 | value | 5100 |  |
| LPP-PDU-Definitions | nrMaxTRPsPerFreq-r16 | value | 5101 |  |
| LPP-PDU-Definitions | nrMaxTRPsPerFreq-1-r16 | value | 5102 |  |
| LPP-PDU-Definitions | maxSimultaneousBands-r16 | value | 5103 |  |
| LPP-PDU-Definitions | maxBandComb-r16 | value | 5105 |  |
| LPP-PDU-Definitions | nrMaxConfiguredBands-r16 | value | 5106 |  |
