# SCF - IAC-28 - Identity Proofing (Identity Verification)
Mechanisms exist to verify the identity of a user before issuing authenticators or modifying access permissions.
## Mapped framework controls
### NIST 800-53
- [IA-12](../nist80053/ia-12.md)

## Evidence request list


## Control questions
Does the organization verify the identity of a user before issuing authenticators or modifying access permissions?

## Compliance methods


## Control maturity
### Not performed
There is no evidence of a capability to verify the identity of a user before issuing authenticators or modifying access permissions.

### Performed internally
SP-CMM1 is N/A, since a structured process is required to verify the identity of a user before issuing authenticators or modifying access permissions.

### Planned and tracked
Identification & Authentication (IAC) efforts are requirements-driven and formally governed at a local/regional level, but are not consistent across the organization. CMM Level 2 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- Logical Access Control (LAC) is decentralized (e.g., a localized/regionalized function) and uses non-standardized methods to implement secure and compliant practices.
- IT/cybersecurity personnel identify cybersecurity & data privacy controls that are appropriate to address applicable statutory, regulatory and contractual requirements for logical access control.
- IT personnel:
o	Implement and maintain an Identity & Access Management (IAM) capability for all users to implement “least privileges” Role Based Access Control (RBAC) practices for the management of user, group and system accounts, including privileged accounts.
o	Govern IAM technologies via RBAC to prohibit privileged access by non-organizational users, unless there is an explicit support contract for privileged IT support services.
- Active Directory (AD), or a similar technology, is primarily used to centrally manage identities and permissions with RBAC. Due to technical or business limitations, asset/process owners are empowered to operate a decentralized access control program for their specific systems, applications and services that cannot be integrated into AD.
- IAM controls are primarily administrative in nature (e.g., policies & standards) to manage accounts and permissions.
- Administrative processes exist to require that a notice of proofing be delivered through an out-of-band channel to verify the user's address (physical or digital).

### Well defined
Identification & Authentication (IAC) efforts are standardized across the organization and centrally managed, where technically feasible, to ensure consistency. CMM Level 3 control maturity would reasonably expect all, or at least most, the following criteria to exist:
- An Identity & Access Management (IAM) function, or similar function, centrally manages permissions and implements “least privileges” Role Based Access Control (RBAC) practices for the management of user, group and system accounts, including privileged accounts.
- The Human Resources (HR) department governs personnel management operations and notifies IAM personnel of personnel role changes for RBAC-based provisioning and deprovisioning actions.
- An IT Asset Management (ITAM) function, or similar function, categorizes endpoint devices according to the data the asset stores, transmits and/ or processes and applies the appropriate technology controls to protect the asset and data that conform to industry-recognized standards for hardening (e.g., DISA STIGs, CIS Benchmarks or OEM security guides).
- An IT infrastructure team, or similar function, ensures that statutory, regulatory and contractual cybersecurity & data privacy obligations are addressed to ensure secure configurations are designed, built and maintained.
- Active Directory (AD), or a similar technology, is used to centrally manage identities and permissions. Only by exception due to a technical or business limitation are solutions authorized to operate a decentralized access control program for systems, applications and services.
- Administrative processes exist to require that a notice of proofing be delivered through an out-of-band channel to verify the user's address (physical or digital).

### Quantitatively controlled
See SP-CMM3. SP-CMM4 is N/A, since a quantitatively-controlled process is not necessary to verify the identity of a user before issuing authenticators or modifying access permissions.

### Continuously improving
See SP-CMM4. SP-CMM5 is N/A, since a continuously-improving process is not necessary to verify the identity of a user before issuing authenticators or modifying access permissions.
