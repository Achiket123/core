# SCF - NET-10.2 - Secure Name / Address Resolution Service (Recursive or Caching Resolver)
Mechanisms exist to perform data origin authentication and data integrity verification on the Domain Name Service (DNS) resolution responses received from authoritative sources when requested by client systems.
## Mapped framework controls
### NIST 800-53
- [SC-21](../nist80053/sc-21.md)

## Evidence request list


## Control questions
Does the organization perform data origin authentication and data integrity verification on the Domain Name Service (DNS) resolution responses received from authoritative sources when requested by client systems?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to perform data origin authentication and data integrity verification on the Domain Name Service (DNS) resolution responses received from authoritative sources when requested by client systems.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to perform data origin authentication and data integrity verification on the Domain Name Service (DNS) resolution responses received from authoritative sources when requested by client systems.

### Planned and tracked
Network Security (NET) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Network security management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for network security management.
- IT personnel define secure networking practices to protect the confidentiality, integrity, availability and safety of the organization's technology assets, data and network(s).
- Administrative processes and technologies focus on protecting High Value Assets (HVAs), including environments where sensitive/regulated data is stored, transmitted and processed.
- Administrative processes are used to configure boundary devices (e.g., firewalls, routers, etc.) to deny network traffic by default and allow network traffic by exception (e.g., deny all, permit by exception).
- Network segmentation exists to implement separate network addresses (e.g., different subnets) to connect systems in different security domains (e.g., sensitive/regulated data environments).
- Administrative processes ensure systems that collectively provide Domain Name Service (DNS) resolution service for are fault-tolerant and implement internal/external role separation.

### Well defined
Network Security (NET) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- A Technology Infrastructure team, or similar function, defines centrally-managed network security controls for implementation across the enterprise.
- Secure engineering principles are used to design and implement network security controls (e.g., industry-recognized secure practices) to enforce the concepts of least privilege and least functionality at the network level.
- IT/cybersecurity architects work with the Technology Infrastructure team to implement a “layered defense” network architecture that provides a defense-in-depth approach for redundancy and risk reduction for network-based security controls, including wired and wireless networking.
- Administrative processes and technologies configure boundary devices (e.g., firewalls, routers, etc.) to deny network traffic by default and allow network traffic by exception (e.g., deny all, permit by exception).
- Technologies automate the Access Control Lists (ACLs) and similar rulesets review process to identify security issues and/ or misconfigurations.
- Network segmentation exists to implement separate network addresses (e.g., different subnets) to connect systems in different security domains (e.g., sensitive/regulated data environments).

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to perform data origin authentication and data integrity verification on the Domain Name Service (DNS) resolution responses received from authoritative sources when requested by client systems.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to perform data origin authentication and data integrity verification on the Domain Name Service (DNS) resolution responses received from authoritative sources when requested by client systems.
