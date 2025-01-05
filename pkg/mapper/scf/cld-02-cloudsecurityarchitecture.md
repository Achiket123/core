# SCF - CLD-02 - Cloud Security Architecture
Mechanisms exist to ensure the cloud security architecture supports the organization's technology strategy to securely design, configure and maintain cloud employments.
## Mapped framework controls
### ISO 27002
- [A.5.23](../iso27002/a-5.md#a523)

## Evidence request list
E-TDA-09

## Control questions
Does the organization ensure the cloud security architecture supports the organization's technology strategy to securely design, configure and maintain cloud employments?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to ensure the cloud security architecture supports the organization's technology strategy to securely design, configure and maintain cloud employments.

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
- Technologies exist to support:
o	A secure infrastructure, including a managed security zone to house cybersecurity & data privacy tools.
o	A standardized virtualization format.
o	Cloud access points, including a managed security zone with
o	Data handling & portability, including a managed security zone to house cybersecurity & data privacy tools
o	Integrity of multi-tenant CSP assets, including a managed security zone to house cybersecurity & data privacy tools
o	Integrity of VM images, including a managed security zone to house cybersecurity & data privacy tools.
o	Processing and storage of service location, including a managed security zone to house cybersecurity & data privacy tools.

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
- IT architects, in conjunction with cybersecurity architects, ensure multi-tenant CSP assets (physical and virtual) are designed and governed such that provider and customer (tenant) user access is appropriately segmented from other tenant users.
- IT architects, in conjunction with cybersecurity architects, ensure CSPs use secure protocols for the import, export and management of data in cloud-based services.
- IT architects, in conjunction with cybersecurity architects, implement a dedicated subnet to host security-specific technologies on all cloud instances, where technically feasible.
- IT infrastructure personnel and Data Protection Officers (DPOs) work with business stakeholders to identify business-critical systems and services, as well as associated sensitive/regulated data, including Personal Data (PD).
- The DPO function oversees the storage, processing and transmission of PD in CSPs.
- An IT Asset Management (ITAM) function, or similar function, governs cloud-based assets leveraging an established Configuration Management Database (CMDB), or similar tool, as the authoritative source of IT assets to provide oversight of purchasing, updating, repairing and disposing of cloud-based assets.
- Formal Change Management (CM) program governs cloud-based systems, applications and services and ensures that no unauthorized changes are made, that all changes are documented, that services are not unnecessarily disrupted and that resources are used efficiently.
- An IT infrastructure team, or similar function, enables the implementation of a Content Delivery Network (CDN) by restricting access to the origin server's IP address to the CDN and an authorized management network.

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to ensure the cloud security architecture supports the organization's technology strategy to securely design, configure and maintain cloud employments.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to ensure the cloud security architecture supports the organization's technology strategy to securely design, configure and maintain cloud employments.
