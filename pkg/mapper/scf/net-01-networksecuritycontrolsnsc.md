# SCF - NET-01 - Network Security Controls (NSC)
Mechanisms exist to develop, govern & update procedures to facilitate the implementation of Network Security Controls (NSC).
## Mapped framework controls
### GDPR
- [Art 32.1](../gdpr/art32.md#Article-321)
- [Art 32.2](../gdpr/art32.md#Article-322)

### ISO 27002
- [A.5.14](../iso27002/a-5.md#a514)
- [A.8.12](../iso27002/a-8.md#a812)
- [A.8.20](../iso27002/a-8.md#a820)
- [A.8.21](../iso27002/a-8.md#a821)

### NIST 800-53
- [SC-1](../nist80053/sc-1.md)

### SOC 2
- [CC6.1-POF5](../soc2/cc61-pof5.md)
- [CC6.1](../soc2/cc61.md)
- [CC6.6-POF1](../soc2/cc66-pof1.md)
- [CC6.6-POF2](../soc2/cc66-pof2.md)
- [CC6.6-POF3](../soc2/cc66-pof3.md)
- [CC6.6-POF4](../soc2/cc66-pof4.md)
- [CC6.6](../soc2/cc66.md)

## Evidence request list


## Control questions
Does the organization develop, govern & update procedures to facilitate the implementation of Network Security Controls (NSC)?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to develop, govern & update procedures to facilitate the implementation of Network Security Controls (NSC).

### Performed internally
Network Security (NET) efforts are ad hoc and inconsistent. CMM Level 1 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- IT personnel use an informal process to design, build and maintain secure networks for test, development, staging and production environments, including the implementation of appropriate cybersecurity & data privacy controls.
- Administrative processes are used to configure boundary devices (e.g., firewalls, routers, etc.) to deny network traffic by default and allow network traffic by exception (e.g., deny all, permit by exception).
- Network monitoring is primarily reactive in nature.

### Planned and tracked
Network Security (NET) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Network security management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for network security management.
- IT personnel define secure networking practices to protect the confidentiality, integrity, availability and safety of the organization's technology assets, data and network(s).
- Administrative processes and technologies focus on protecting High Value Assets (HVAs), including environments where sensitive/regulated data is stored, transmitted and processed.
- Administrative processes are used to configure boundary devices (e.g., firewalls, routers, etc.) to deny network traffic by default and allow network traffic by exception (e.g., deny all, permit by exception).
- Network segmentation exists to implement separate network addresses (e.g., different subnets) to connect systems in different security domains (e.g., sensitive/regulated data environments).

### Well defined
Network Security (NET) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- A Technology Infrastructure team, or similar function, defines centrally-managed network security controls for implementation across the enterprise.
- Secure engineering principles are used to design and implement network security controls (e.g., industry-recognized secure practices) to enforce the concepts of least privilege and least functionality at the network level.
- IT/cybersecurity architects work with the Technology Infrastructure team to implement a “layered defense” network architecture that provides a defense-in-depth approach for redundancy and risk reduction for network-based security controls, including wired and wireless networking.
- Administrative processes and technologies configure boundary devices (e.g., firewalls, routers, etc.) to deny network traffic by default and allow network traffic by exception (e.g., deny all, permit by exception).
- Technologies automate the Access Control Lists (ACLs) and similar rulesets review process to identify security issues and/ or misconfigurations.
- Network segmentation exists to implement separate network addresses (e.g., different subnets) to connect systems in different security domains (e.g., sensitive/regulated data environments).

### Quantitatively controlled
Network Security (NET) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to develop, govern & update procedures to facilitate the implementation of Network Security Controls (NSC).
