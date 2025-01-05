# SCF - CHG-04.1 - Automated Access Enforcement / Auditing
Mechanisms exist to perform after-the-fact reviews of configuration change logs to discover any unauthorized changes.
## Mapped framework controls
### SOC 2
- [CC8.1-POF10](../soc2/cc81-pof10.md)
- [CC8.1-POF11](../soc2/cc81-pof11.md)

## Evidence request list


## Control questions
Does the organization perform after-the-fact reviews of configuration change logs to discover any unauthorized changes?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to perform after-the-fact reviews of configuration change logs to discover any unauthorized changes.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to perform after-the-fact reviews of configuration change logs to discover any unauthorized changes.

### Planned and tracked
SP-CMM2 is N/A, since a well-defined process is required to perform after-the-fact reviews of configuration change logs to discover any unauthorized changes.

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
Change Management (CHG) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to perform after-the-fact reviews of configuration change logs to discover any unauthorized changes.
