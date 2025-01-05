# SCF - CLD-11 - Cloud Access Security Broker (CASB)
Mechanisms exist to utilize a Cloud Access Security Broker (CASB), or similar technology, to provide boundary protection and monitoring functions that both provide access to the cloud and protect the organization from misuse of cloud resources.
## Mapped framework controls
### SOC 2
- [CC6.1-POF5](../soc2/cc61-pof5.md)

## Evidence request list


## Control questions
Does the organization utilize a Cloud Access Security Broker (CASB), or similar technology, to provide boundary protection and monitoring functions that both provide access to the cloud and protect the organization from misuse of cloud resources?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to utilize a Cloud Access Security Broker (CASB), or similar technology, to provide boundary protection and monitoring functions that both provide access to the cloud and protect the organization from misuse of cloud resources.

### Performed internally
Cloud Security (CLD) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Cloud-based technologies are governed no differently from on-premise network assets (e.g., cloud-based technology is viewed as an extension of the corporate network).
- A Shared Responsibility Matrix (SRM), or similar Customer Responsibility Matrix (CRM), is documented for each Cloud Service Providers (CSPs) instance that takes into account differences between Software as a Service (SaaS), Platform as a Service (PaaS) and Infrastructure as a Service (IaaS) methodologies.

### Planned and tracked
SP-CMM2 is N/A, since a well-defined process is required to prevent "side channel attacks" when using a Content Delivery Network (CDN) by restricting access to the origin server's IP address to the CDN and an authorized management network.

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
- IT architects, in conjunction with cybersecurity architects, use Cloud Access Points (CAPs) to provide boundary protection and monitoring functions that control access to the cloud and protect the organization as well.

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to utilize a Cloud Access Security Broker (CASB), or similar technology, to provide boundary protection and monitoring functions that both provide access to the cloud and protect the organization from misuse of cloud resources.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to utilize a Cloud Access Security Broker (CASB), or similar technology, to provide boundary protection and monitoring functions that both provide access to the cloud and protect the organization from misuse of cloud resources.
