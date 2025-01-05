# SCF - MDM-05 - Remote Purging
Mechanisms exist to remotely purge selected information from mobile devices.
## Mapped framework controls
### ISO 27002
- [A.8.1](../iso27002/a-8.md#a81)

## Evidence request list


## Control questions
Does the organization remotely purge selected information from mobile devices?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to remotely purge selected information from mobile devices.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to remotely purge selected information from mobile devices.

### Planned and tracked
Mobile Device Management (MDM) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Mobile device management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel:
o	Identify cybersecurity & data privacy controls to address applicable statutory, regulatory and contractual requirements for Mobile Device Management (MDM).
o	Implement and maintain a MDM capability for all mobile devices in use at the organization.
- Organization-owned mobile devices are configured to protect data with the strength and integrity commensurate with the classification or sensitivity of the information stored on the device and mostly conform to industry-recognized standards for hardening (e.g., DISA STIGs, CIS Benchmarks or OEM security guides), including cryptographic protections for sensitive/regulated data.
- MDM software is used to restrict the data that is stored/processed/transmitted on organization-owned and/ or applicable Bring Your Own Device (BYOD) (e.g., personal devices).
- Administrative processes and technologies exist to remotely purge selected information from mobile devices.

### Well defined
Mobile Device Management (MDM) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- An Identity & Access Management (IAM) function, or similar function, performs the implementation of access controls for mobile devices that restricts the connectivity of mobile devices from communicating with systems, applications and services.
- Organization-owned mobile devices are configured to protect data with the strength and integrity commensurate with the classification or sensitivity of the information stored on the device, and conform to industry-recognized standards for hardening (e.g., DISA STIGs, CIS Benchmarks or OEM security guides), including cryptographic protections for sensitive/regulated data.
- MDM software is used to restrict the data that is stored/processed/transmitted on organization-owned and/ or applicable Bring Your Own Device (BYOD) (e.g., personal devices) across the entire organization.
- MDM enforces a separate device workspace on applicable mobile devices to separate work-related and personal-related applications and data.
- Technologies are configured to use cryptographic mechanisms to protect the confidentiality and integrity of information on mobile devices through full-device or container encryption.

### Quantitatively controlled
Mobile Device Management (MDM) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
Mobile Device Management (MDM) efforts are “world-class” capabilities that leverage predictive analysis (e.g., machine learning, AI, etc.). In addition to CMM Level 4 criteria, CMM Level 5 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Stakeholders make time-sensitive decisions to support operational efficiency, which may include automated remediation actions.
- Based on predictive analysis, process improvements are implemented according to “continuous improvement” practices that affect process changes.
