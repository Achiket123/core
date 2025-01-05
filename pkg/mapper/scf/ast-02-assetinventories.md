# SCF - AST-02 - Asset Inventories
Mechanisms exist to perform inventories of technology assets that:
 - Accurately reflects the current systems, applications and services in use;
 - Identifies authorized software products, including business justification details;
 - Is at the level of granularity deemed necessary for tracking and reporting;
 - Includes organization-defined information deemed necessary to achieve effective property accountability; and
 - Is available for review and audit by designated organizational personnel.
## Mapped framework controls
### ISO 27002
- [A.5.9](../iso27002/a-5.md#a59)

### NIST 800-53
- [CM-8](../nist80053/cm-8.md)

### SOC 2
- [CC2.1-POF6](../soc2/cc21-pof6.md)
- [CC2.1-POF9](../soc2/cc21-pof9.md)
- [CC6.1-POF1](../soc2/cc61-pof1.md)

## Evidence request list
E-AST-04
E-AST-05
E-AST-07

## Control questions
Does the organization perform inventories of technology assets that:
 •Accurately reflects the current systems, applications and services in use;
 •Identifies authorized software products, including business justification details;
 •Is at the level of granularity deemed necessary for tracking and reporting;
 •Includes organization-defined information deemed necessary to achieve effective property accountability; and
 •Is available for review and audit by designated organizational personnel?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to perform inventories of technology assets that:
 •Accurately reflects the current systems, applications and services in use;
 •Identifies authorized software products, including business justification details;
 •Is at the level of granularity deemed necessary for tracking and reporting;
 •Includes organization-defined information deemed necessary to achieve effective property accountability; and
 •Is available for review and audit by designated organizational personnel.

### Performed internally
Asset Management (AST) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Asset management is informally assigned as an additional duty to existing IT/cybersecurity personnel.
- Asset inventories are performed in an ad hoc manner.
- Software licensing is tracked as part of IT asset inventories.
- Data process owners maintain limited network diagrams to document the flow of sensitive/regulated data that is specific to their initiative.
- IT personnel work with data/process owners to help ensure secure practices are implemented throughout the System Development Lifecycle (SDLC) for all high-value projects.
- Inventories are manual (e.g., spreadsheets).
- Assets are assigned owners and are documented.
- No structured process exists to review or share the results of the inventories.

### Planned and tracked
Asset Management (AST) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Asset management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for asset management.
- Administrative processes and technologies focus on protecting High Value Assets (HVAs), including environments where sensitive/regulated data is stored, transmitted and processed.
- Asset management is formally assigned as an additional duty to existing IT/cybersecurity personnel.
- Technology assets are categorized according to data classification and business criticality.
- Inventories cover technology assets in scope for statutory, regulatory and/ or contractual compliance, which includes both physical and virtual assets.
- Software licensing is tracked as part of IT asset inventories.
- Users are educated on their responsibilities to protect technology assets assigned to them or under their supervision.
- IT/cybersecurity personnel maintain network diagrams to document the flow of sensitive/regulated data across the network.
- Maintenance of asset inventory is performed at least annually.
- Inventory of physical technology assets are assigned to individual users or teams and covers common devices (e.g., laptops, workstations and servers).
- Inventories may be manual (e.g., spreadsheets) or automated.
- No structured process exists to review or share the results of the inventories.
- Annual IT asset inventories validate or update stakeholders /owners.

### Well defined
Asset Management (AST) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- An IT Asset Management (ITAM) function, or similar function, governs asset management to help ensure compliance with requirements for asset management.
- An ITAM function, or similar function, maintains an inventory of IT assets, covering both physical and virtual assets, as well as centrally managed asset ownership assignments.
- Technology assets and data are categorized according to data classification and business criticality criteria.
- A Cybersecurity Supply Chain Risk Management (C-SCRM) function oversees supply chain risks including the removal and prevention of certain technology services and/ or equipment designated as supply chain threats by a statutory or regulatory body.
- Data/process owners document where sensitive/regulated data is stored, transmitted and processed, generating Data Flow Diagrams (DFDs) and network diagrams to document the flow of data.
- Quarterly IT asset inventories are reviewed and shared with appropriate stakeholders.
- Inventories are predominately automated, but may have some manual components (e.g., cloud-based assets that are out of scope for automated inventory scans).
- Inventories processes include Indicators of Compromise (IoC) to identify evidence of physical tampering.
- Inventory scans are configured to be recurring, based on ITAM tool configuration settings.
- Annual IT asset inventories validate or update stakeholders / owners.
- A Software Asset Management (SAM) solution is used to centrally manage deployed software.
- An ITAM function, or similar function, conducts ongoing “technical debt” reviews of hardware and software technologies to remediate outdated and/ or unsupported technologies.

### Quantitatively controlled
Asset Management (AST) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement.
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
Asset Management (AST) efforts are “world-class” capabilities that leverage predictive analysis (e.g., machine learning, AI, etc.). In addition to CMM Level 4 criteria, CMM Level 5 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Stakeholders make time-sensitive decisions to support operational efficiency, which may include automated remediation actions.
- Based on predictive analysis, process improvements are implemented according to “continuous improvement” practices that affect process changes.
