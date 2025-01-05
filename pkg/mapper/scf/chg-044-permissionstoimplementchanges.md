# SCF - CHG-04.4 - Permissions To Implement Changes
Mechanisms exist to limit operational privileges for implementing changes.
## Mapped framework controls
### SOC 2
- [CC8.1-POF9](../soc2/cc81-pof9.md)

## Evidence request list


## Control questions
Does the organization limit operational privileges for implementing changes?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to limit operational privileges for implementing changes.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to limit operational privileges for implementing changes.

### Planned and tracked
Change Management (CHG) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Change management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for change management.
- Changes are tracked through a centralized technology solution to submit, review, approve and assign Requests for Change (RFC).
- A Change Advisory Board (CAB), or similar function, exists to govern changes to systems, applications and services to ensure their stability, reliability and predictability.
- A CAB, or similar function, reviews RFCs for cybersecurity & data privacy ramifications.
- A CAB, or similar function, notifies stakeholders to ensure awareness of the impact of proposed changes.
- Logical Access Control (LAC) limits the ability of non-administrators from making unauthorized configuration changes to systems, applications and services.
- Cybersecurity controls are tested after a change is implemented to ensure cybersecurity controls are operating properly.

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

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to limit operational privileges for implementing changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to limit operational privileges for implementing changes.
