# SCF - DCH-10.1 - Limitations on Use
Mechanisms exist to restrict the use and distribution of sensitive / regulated data.
## Mapped framework controls
### ISO 27002
- [A.7.10](../iso27002/a-7.md#a710)

## Evidence request list


## Control questions
Does the organization restrict the use and distribution of sensitive / regulated data?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to restrict the use and distribution of sensitive / regulated data.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to restrict the use and distribution of sensitive / regulated data.

### Planned and tracked
Data Classification & Handling (DCH) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Data management is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for data management.
- Data protection controls are primarily administrative and preventative in nature (e.g., policies & standards) to classify, protect and dispose of systems and data, including storage media.
- A data classification process exists to identify categories of data and specific protection requirements.
- A data retention process exists and is a manual process to govern.
- Data/process owners:
o	Document where sensitive/regulated data is stored, transmitted and processed to identify data repositories and data flows.
o	Create and maintain Data Flow Diagrams (DFDs) and network diagrams.
o	Are expected to take the initiative to work with Data Protection Officers (DPOs) to ensure applicable statutory, regulatory and contractual obligations are properly addressed, including the storage, transmission and processing of sensitive/regulated data
- A manual data retention process exists.
- Content filtering blocks users from performing ad hoc file transfers through unapproved file transfer services (e.g., Box, Dropbox, Google Drive, etc.).
- Mobile Device Management (MDM) software is used to restrict and protect the data that resides on mobile devices.
- Physical controls, administrative processes and technologies focus on protecting High Value Assets (HVAs), including environments where sensitive/regulated data is stored, transmitted and processed.
- Administrative means (e.g., policies and standards) dictate:
o	Geolocation requirements for sensitive/regulated data types, including the transfer of data to third-countries or international organizations.
o	Requirements for minimizing data collection to what is necessary for business purposes.
o	Requirements for limiting the use of sensitive/regulated data in testing, training and research.


### Well defined
Data Classification & Handling (DCH) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- A Governance, Risk & Compliance (GRC) function, or similar function, assists users in making information sharing decisions to ensure data is appropriately protected, regardless of where or how it is stored, processed and/ or transmitted.
- A data classification process exists to identify categories of data and specific protection requirements.
- A data retention process exists to protect archived data in accordance with applicable statutory, regulatory and contractual obligations.
- Data/process owners:
o	Are expected to take the initiative to work with Data Protection Officers (DPOs) to ensure applicable statutory, regulatory and contractual obligations are properly addressed, including the storage, transmission and processing of sensitive/regulated data.
o	Maintain decentralized inventory logs of all sensitive/regulated media and update sensitive/regulated media inventories at least annually.
o	Create and maintain Data Flow Diagrams (DFDs) and network diagrams.
o	Document where sensitive/regulated data is stored, transmitted and processed in order to document data repositories and data flows.
- A Data Protection Impact Assessment (DPIA) is used to help ensure the protection of sensitive/regulated data processed, stored or transmitted on internal or external systems, in order to implement cybersecurity & data privacy controls in accordance with applicable statutory, regulatory and contractual obligations.
- Human Resources (HR), documents formal “rules of behavior” as an employment requirement that stipulates acceptable and unacceptable practices pertaining to sensitive/regulated data handling.
- Data Loss Prevention (DLP), or similar content filtering capabilities, blocks users from performing ad hoc file transfers through unapproved file transfer services (e.g., Box, Dropbox, Google Drive, etc.).
- Mobile Device Management (MDM) software is used to restrict and protect the data that resides on mobile devices.
- Administrative processes and technologies:
o	Identify data classification types to ensure adequate cybersecurity & data privacy controls are in place to protect organizational information and individual data privacy.
o	Identify and document the location of information on which the information resides.
o	Restrict and govern the transfer of data to third-countries or international organizations.
o	Limit the disclosure of data to authorized parties.
o	Mark media in accordance with data protection requirements so that personnel are alerted to distribution limitations, handling caveats and applicable security requirements.
o	Prohibit “rogue instances” where unapproved third parties are engaged to store, process or transmit data, including budget reviews and firewall connection authorizations.
o	Protect and control digital and non-digital media during transport outside of controlled areas using appropriate security measures.
o	Govern the use of personal devices (e.g., Bring Your Own Device (BYOD)) as part of acceptable and unacceptable behaviors.
o	Dictate requirements for minimizing data collection to what is necessary for business purposes.
o	Dictate requirements for limiting the use of sensitive/regulated data in testing, training and research.
- Administrative processes and technologies:
o	Collect Personal data (PD) directly from the individual.
o	Correct Personal data (PD) that is inaccurate or outdated, incorrectly determined regarding impact, or incorrectly de-identified.
o	De-identify the dataset up on collection by not collecting Personal data (PD).
o	Govern how data is reclassified due to changing business/technical requirements to ensure the integrity of data classification is upheld through the data lifecycle.
o	Limit Personal data (PD) being processed in the information lifecycle to elements identified in the Data Protection Impact Assessment (DPIA).
o	Refrain from archiving Personal data (PD) elements if those elements in a dataset will not be needed after the dataset is archived.
o	Remove Personal data (PD) elements from a dataset prior to its release if those elements in the dataset do not need to be part of the data release.
o	Remove Personal data (PD) from datasets.
o	Limit Personal data (PD) being processed in the information lifecycle to elements identified in the DPIA.
o	Identify custodians throughout the transport of system media.
o	Minimize the use of Personal data (PD) for research, testing or training, in accordance with the DPIA.
o	Minimize the use of Personal data (PD) for research, testing, or training, in accordance with the Data Protection Impact Assessment (DPIA).
o	Perform a motivated intruder test on the de-identified dataset to determine if the identified data remains or if the de-identified data can be re-identified.

### Quantitatively controlled
Data Classification & Handling (DCH) efforts are metrics driven and provide sufficient management insight (based on a quantitative understanding of process capabilities) to predict optimal performance, ensure continued operations and identify areas for improvement. In addition to CMM Level 3 criteria, CMM Level 4 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Metrics reporting includes quantitative analysis of Key Performance Indicators (KPIs).
- Metrics reporting includes quantitative analysis of Key Risk Indicators (KRIs).
- Scope of metrics, KPIs and KRIs covers organization-wide cybersecurity & data privacy controls, including functions performed by third-parties.
- Organizational leadership maintains a formal process to objectively review and respond to metrics, KPIs and KRIs (e.g., monthly or quarterly review).
- Based on metrics analysis, process improvement recommendations are submitted for review and are handled in accordance with change control processes.
- Both business and technical stakeholders are involved in reviewing and approving proposed changes.

### Continuously improving
Data Classification & Handling (DCH) efforts are “world-class” capabilities that leverage predictive analysis (e.g., machine learning, AI, etc.). In addition to CMM Level 4 criteria, CMM Level 5 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Stakeholders make time-sensitive decisions to support operational efficiency, which may include automated remediation actions.
- Based on predictive analysis, process improvements are implemented according to “continuous improvement” practices that affect process changes.
