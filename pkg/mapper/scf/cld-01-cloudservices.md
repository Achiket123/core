# SCF - CLD-01 - Cloud Services
Mechanisms exist to facilitate the implementation of cloud management controls to ensure cloud instances are secure and in-line with industry practices.
## Mapped framework controls
### GDPR
- [Art 32.1](../gdpr/art32.md#Article-321)
- [Art 32.2](../gdpr/art32.md#Article-322)

### ISO 27002
- [A.5.23](../iso27002/a-5.md#a523)

### SOC 2
- [CC6.1-POF5](../soc2/cc61-pof5.md)

## Evidence request list
E-AST-06

## Control questions
Does the organization facilitate the implementation of cloud management controls to ensure cloud instances are secure and in-line with industry practices?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to facilitate the implementation of cloud management controls to ensure cloud instances are secure and in-line with industry practices.

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
- The Chief Information Security Officer (CISO), or similar function with technical competence to address cybersecurity concerns, analyzes the organization's business strategy to determine prioritized and authoritative guidance for cloud security practices.
- The CISO, or similar function, develops a security-focused Concept of Operations (CONOPS) that documents management, operational and technical measures to apply defense-in-depth techniques across the organization to ensure that cloud security is incorporated.
- A Governance, Risk & Compliance (GRC) function, or similar function, provides governance oversight for the implementation of applicable statutory, regulatory and contractual cybersecurity & data privacy controls to protect the confidentiality, integrity, availability and safety of the organization's applications, systems, services and data to ensure that compliance requirements for cloud security are identified and documented.
- A Chief Information Officer (CIO), or similar function, defines the authoritative architecture for use with on-premise, cloud-native and hybrid models, providing governance oversight for operations planning, deployment and maintenance of cloud-based technology assets supporting cybersecurity and data protection requirements.
- A Chief Technology Officer (CTO), or similar function, aligns with the CIO’s architectural model to evaluate and implement new cloud-based technologies.
- A steering committee is formally established to provide executive oversight of the cybersecurity & data privacy program, including cloud security, as well as establish a clear and authoritative accountability structure for cloud security operations.
- Roles and associated responsibilities for governing cloud instances, including provisioning, maintaining and deprovisioning, are formally assigned.
- A Shared Responsibility Matrix (SRM), or similar Customer Responsibility Matrix (CRM), is documented for each Cloud Service Providers (CSPs) instance that takes into account differences between Software as a Service (SaaS), Platform as a Service (PaaS) and Infrastructure as a Service (IaaS) methodologies.
- IT architects, in conjunction with cybersecurity architects:
o	Ensure the cloud security architecture supports the organization's technology strategy to securely design, configure and maintain cloud employments.
o	Ensure multi-tenant CSP assets (physical and virtual) are designed and governed such that provider and customer (tenant) user access is appropriately segmented from other tenant users.
o	Ensure CSPs use secure protocols for the import, export and management of data in cloud-based services.
o	Implement a dedicated subnet to host security-specific technologies on all cloud instances, where technically feasible.
- A Change Advisory Board (CAB), or similar function, governs changes to cloud-based systems, applications and services to ensure their stability, reliability and predictability.
- CAB review processes identify and prevent use of unapproved CSPs.
- A dedicated IT infrastructure team, or similar function, enables the implementation of cloud management controls to ensure cloud instances are both secure and compliant, leveraging industry-recognized secure practices that are CSP-specific.
- Cybersecurity and data privacy requirements are identified and documented for each CSP instance to address sensitive/regulated data processing, storing and/ or transmitting and provide restrictions on data processing and storage locations.
- A Data Protection Impact Assessment (DPIA) is used to help ensure the protection of sensitive/regulated data processed, stored or transmitted on external systems.

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to facilitate the implementation of cloud management controls to ensure cloud instances are secure and in-line with industry practices.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to facilitate the implementation of cloud management controls to ensure cloud instances are secure and in-line with industry practices.
