# UE GNSS live-test status

Do not run a UE GNSS positioning trial against this increment. The SLs server
now decodes the bounded LPP payload-type 0 carrier and routes it to the LPP
procedure, but it does not start a GNSS method, obtain GNSS capabilities or
assistance, or derive an estimate. Static simulation must remain disabled for
any future real-positioning test.

The MME-side prerequisites for a future test are an established PPID 29 SCTP
association, stream 0, SLg connectivity, NAS security, and an LTE LPP/GNSS
capable UE. A valid future pass requires a RequestCapabilities/ProvideCapabilities
exchange, a RequestLocationInformation/ProvideLocationInformation exchange, a
UE-derived decoded estimate, successful Location Response, and zero remaining
session state. NAS LPP may be encrypted over S1AP and not visible in captures.
