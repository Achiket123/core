# SCF - TPM-01 - Third-Party Management
Mechanisms exist to facilitate the implementation of third-party management controls.
## Mapped framework controls
### GDPR
- [Art 28.10](../gdpr/art28.md#Article-2810)
- [Art 28.1](../gdpr/art28.md#Article-281)
- [Art 28.2](../gdpr/art28.md#Article-282)
- [Art 28.3](../gdpr/art28.md#Article-283)
- [Art 28.4](../gdpr/art28.md#Article-284)
- [Art 28.5](../gdpr/art28.md#Article-285)
- [Art 28.6](../gdpr/art28.md#Article-286)
- [Art 28.9](../gdpr/art28.md#Article-289)
- [Art 32.1](../gdpr/art32.md#Article-321)
- [Art 32.2](../gdpr/art32.md#Article-322)

### ISO 27002
- [A.5.19](../iso27002/a-5.md#a519)
- [A.5.20](../iso27002/a-5.md#a520)
- [A.8.30](../iso27002/a-8.md#a830)

### NIST 800-53
- [SA-4](../nist80053/sa-4.md)
- [SR-1](../nist80053/sr-1.md)

### SOC 2
- [CC1.1-POF5](../soc2/cc11-pof5.md)
- [CC1.4-POF2](../soc2/cc14-pof2.md)
- [CC1.4-POF3](../soc2/cc14-pof3.md)
- [CC2.3-POF10](../soc2/cc23-pof10.md)
- [CC2.3-POF12](../soc2/cc23-pof12.md)
- [CC2.3-POF9](../soc2/cc23-pof9.md)
- [CC3.3](../soc2/cc33.md)
- [CC3.4-POF5](../soc2/cc34-pof5.md)
- [CC9.1](../soc2/cc91.md)
- [CC9.2-POF10](../soc2/cc92-pof10.md)
- [CC9.2-POF11](../soc2/cc92-pof11.md)
- [CC9.2-POF12](../soc2/cc92-pof12.md)
- [CC9.2-POF1](../soc2/cc92-pof1.md)
- [CC9.2-POF2](../soc2/cc92-pof2.md)
- [CC9.2-POF3](../soc2/cc92-pof3.md)
- [CC9.2-POF4](../soc2/cc92-pof4.md)
- [CC9.2-POF5](../soc2/cc92-pof5.md)
- [CC9.2-POF6](../soc2/cc92-pof6.md)
- [CC9.2-POF7](../soc2/cc92-pof7.md)
- [CC9.2-POF8](../soc2/cc92-pof8.md)
- [CC9.2-POF9](../soc2/cc92-pof9.md)
- [CC9.2](../soc2/cc92.md)

## Evidence request list
E-TPM-03

## Control questions
Does the organization facilitate the implementation of third-party management controls?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to facilitate the implementation of third-party management controls.

### Performed internally
Third-Party Management (TPM) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Third-party management is decentralized.
- IT personnel use an informal process to govern third-party service providers.
- IT personnel work with data/process owners to help ensure secure practices are implemented throughout the System Development Lifecycle (SDLC) for all high-value projects.
- Project management is decentralized and generally lacks formal project management managers or broader oversight.

### Planned and tracked
Third-Party Management (TPM) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Third-party management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls to address applicable statutory, regulatory and contractual requirements for third-party management.
- A procurement function maintains a list of all active Third-Party Service Providers (TSP), including pertinent contract information that will assist in a risk assessment.
- A Shared Responsibility Matrix (SRM) is documented for every TSP that directly or indirectly affects sensitive/regulated data.
- Procurement contracts:
o	Require TSP to follow secure engineering practices as part of a broader Cybersecurity Supply Chain Risk Management (C-SCRM) initiative.
o	Contain "break clauses" in all TSP contracts to enable penalty-free, early termination of a contract for cause, based on the TSP's cybersecurity and/ or data privacy practices deficiency(ies).


### Well defined
Third-Party Management (TPM) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- The Chief Information Security Officer (CISO), or similar function with technical competence to address cybersecurity concerns, analyzes the organization's business strategy to determine prioritized and authoritative guidance for third-party management practices.
- The CISO, or similar function, develops a security-focused Concept of Operations (CONOPS) that documents management, operational and technical measures to apply defense-in-depth techniques across the enterprise for third-party management.
- A steering committee is formally established to provide executive oversight of the cybersecurity & data privacy program, including third-party management.
- Procurement contracts and layered defenses provide safeguards to limit harm from potential adversaries who identify and target the organization's supply chain.
- A Governance, Risk & Compliance (GRC) function, or similar function;
o	provides governance oversight for the implementation of applicable statutory, regulatory and contractual cybersecurity & data privacy controls to protect the confidentiality, integrity, availability and safety of the organization's applications, systems, services and data with regards to third-party management.
o	Operates the Cybersecurity Supply Chain Risk Management (C-SCRM) program to identify and mitigate supply chain-related risks and threats.
o	Evaluates risks associated with weaknesses or deficiencies in supply chain elements identified during first and/ or third-party reviews.
o	Enables the implementation of third-party management controls.
o	Ensures the Information Assurance Program (IAP) evaluates applicable cybersecurity & data privacy controls as part of “business as usual” pre-production testing.
- A procurement team, or similar function:
o	Maintains a list of all active Third-Party Service Providers (TSP), including pertinent contract information that will assist in a risk assessment.
o	Requires TSP to follow secure engineering practices as part of a broader Cybersecurity Supply Chain Risk Management (C-SCRM) initiative.
o	Includes "break clauses" in all TSP contracts to enable penalty-free, early termination of a contract for cause, based on the TSP's cybersecurity and/ or data privacy practices deficiency(ies).
o	Controls changes to services by suppliers, taking into account the criticality of business information, systems and processes that are in scope by the third-party.
o	Requires a risk assessment prior to the acquisition or outsourcing of technology-related services.
o	Monitors, regularly reviews and audits supplier service delivery for compliance with established contract agreements.
o	Uses tailored acquisition strategies, contract tools and procurement methods for the purchase of unique systems, system components or services.
- A Shared Responsibility Matrix (SRM) is documented for every TSP that directly or indirectly affects sensitive/regulated data.

### Quantitatively controlled
Third-Party Management (TPM) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to facilitate the implementation of third-party management controls.
