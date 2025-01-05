# SCF - CLD-09 - Geolocation Requirements for Processing, Storage and Service Locations
Mechanisms exist to control the location of cloud processing/storage based on business requirements that includes statutory, regulatory and contractual obligations.
## Mapped framework controls
### ISO 27002
- [A.5.23](../iso27002/a-5.md#a523)

### SOC 2
- [CC2.1-POF9](../soc2/cc21-pof9.md)

## Evidence request list
E-AST-06
E-AST-23

## Control questions
Does the organization control the location of cloud processing/storage based on business requirements that includes statutory, regulatory and contractual obligations?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to control the location of cloud processing/storage based on business requirements that includes statutory, regulatory and contractual obligations.

### Performed internally
Cloud Security (CLD) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Cloud-based technologies are governed no differently from on-premise network assets (e.g., cloud-based technology is viewed as an extension of the corporate network).
- A Shared Responsibility Matrix (SRM), or similar Customer Responsibility Matrix (CRM), is documented for each Cloud Service Providers (CSPs) instance that takes into account differences between Software as a Service (SaaS), Platform as a Service (PaaS) and Infrastructure as a Service (IaaS) methodologies.

### Planned and tracked
Cloud Security (CLD) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Cloud security management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel:
o	Identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for cloud security management.
o	Use an informal process to govern cloud-specific cybersecurity & data privacy-specific tools.
- A Shared Responsibility Matrix (SRM), or similar Customer Responsibility Matrix (CRM), is documented for each Cloud Service Providers (CSPs) instance that takes into account differences between Software as a Service (SaaS), Platform as a Service (PaaS) and Infrastructure as a Service (IaaS) methodologies.
- IT personnel have a documented architecture for cloud-based technologies to support cybersecurity and data protection requirements.
- Cybersecurity and data privacy requirements are identified and documented for cloud-specific sensitive/regulated data processing, storing and/ or transmitting, including restrictions on data processing and storage locations.

### Well defined
Cloud Security (CLD) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Roles and associated responsibilities for governing cloud instances, including provisioning, maintaining and deprovisioning, are formally assigned.
- A Shared Responsibility Matrix (SRM), or similar Customer Responsibility Matrix (CRM), is documented for each Cloud Service Providers (CSPs) instance that takes into account differences between Software as a Service (SaaS), Platform as a Service (PaaS) and Infrastructure as a Service (IaaS) methodologies.
- IT architects, in conjunction with cybersecurity architects:
o	Ensure the cloud security architecture supports the organization's technology strategy to securely design, configure and maintain cloud employments.
o	Ensure multi-tenant CSP assets (physical and virtual) are designed and governed such that provider and customer (tenant) user access is appropriately segmented from other tenant users.
o	Ensure CSPs use secure protocols for the import, export and management of data in cloud-based services.
o	Implement a dedicated subnet to host security-specific technologies on all cloud instances, where technically feasible.
- A Change Advisory Board (CAB), or similar function:
o	Governs changes to cloud-based systems, applications and services to ensure their stability, reliability and predictability.
o	Reviews processes to identify and prevent use of unapproved CSPs.
- A dedicated IT infrastructure team, or similar function, enables the implementation of cloud management controls to ensure cloud instances are both secure and compliant, leveraging industry-recognized secure practices that are CSP-specific.
- Cybersecurity and data privacy requirements are identified and documented for each CSP instance to address sensitive/regulated data processing, storing and/ or transmitting and provide restrictions on data processing and storage locations.
- A Data Protection Impact Assessment (DPIA) is used to help ensure the protection of sensitive/regulated data processed, stored or transmitted on external systems.

### Quantitatively controlled
Cloud Security (CLD) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
Cloud Security (CLD) efforts are “world-class” capabilities that leverage predictive analysis (e.g., machine learning, AI, etc.). In addition to CMM Level 4 criteria, CMM Level 5 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Stakeholders make time-sensitive decisions to support operational efficiency, which may include automated remediation actions.
- Based on predictive analysis, process improvements are implemented according to “continuous improvement” practices that affect process changes.
