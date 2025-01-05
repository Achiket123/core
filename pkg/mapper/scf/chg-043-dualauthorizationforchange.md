# SCF - CHG-04.3 - Dual Authorization for Change
Mechanisms exist to enforce a two-person rule for implementing changes to critical assets.
## Mapped framework controls
### NIST 800-53
- [AC-5](../nist80053/ac-5.md)

## Evidence request list


## Control questions
Does the organization enforce a two-person rule for implementing changes to critical assets?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to enforce a two-pers on rule for implementing changes to critical assets.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to enforce a two-pers on rule for implementing changes to critical assets.

### Planned and tracked
SP-CMM2 is N/A, since a well-defined process is required to enforce a two-pers on rule for implementing changes to critical assets.

### Well defined
Change Management (CHG) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- An IT Asset Management (ITAM) function, or similar function, ensures compliance with requirements for asset management.
- ITAM leverages a Configuration Management Database (CMDB), or similar tool, as the authoritative source of IT assets.
- Logical Access Control (LAC) is governed to limit the ability of non-administrators from making configuration changes to systems, applications and services.
- A formal Change Management (CM) program ensures that no unauthorized changes are made, that all changes are documented, that services are not disrupted and that resources are used efficiently.
- The CM function has formally defined roles and associated responsibilities.
- Changes are tracked through a centralized technology solution to submit, review, approve and assign Requests for Change (RFC).
- A Change Advisory Board (CAB), or similar function:
o	Exists to govern changes to systems, applications and services to ensure their stability, reliability and predictability.
o	Reviews RFC for cybersecurity & data privacy ramifications.
o	Notifies stakeholders to ensure awareness of the impact of proposed changes.
- IT personnel use dedicated development/test/staging environments to deploy and evaluate changes, wherever technically possible.
- Critical systems are configured to use dual authorization mechanisms requiring the approval of two authorized individuals in order to execute a change.

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to enforce a two-pers on rule for implementing changes to critical assets.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to enforce a two-pers on rule for implementing changes to critical assets.
