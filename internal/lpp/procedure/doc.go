// Package procedure coordinates the bounded LPP request/provide envelope flow.
// An Orchestrator is scoped by its caller to exactly one LPP peer and has no
// transport identity, timer, goroutine, encoder, or positioning payload logic.
// It can carry the bounded ECID capability request selector, typed ECID
// support BIT STRING, and the ECID RequestLocationInformation requested-
// measurements BIT STRING unchanged. Inbound ECID location requests create one
// bounded pending application wait and emit a typed event; an application can
// complete it with the bounded root-only ECID ProvideLocationInformation
// payload. Inbound typed provides are delivered as events. Exact BIT STRING
// content and length are retained, including distinct values such as one-bit 1
// and three-bit 100. It never selects a positioning method, executes a
// measurement, calculates an estimate, or persists peer capabilities.
// It returns messages as actions; callers encode and send them.
package procedure
