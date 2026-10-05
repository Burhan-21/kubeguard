# Web Scraping Policy

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

**Status:** NOT REQUIRED FOR CORE PRODUCT

KubeGuard explicitly does **NOT** require web scraping for its core functionality. All rules, references, and validation logic are deterministically coded and configured.

## Future Automated Doc Collection
If automated document collection or reference synchronization is ever added in the future, it must strictly adhere to the following definitions:

- **Allowed Sources**: Only official CNCF, Kubernetes, OWASP, CIS, or NIST domains.
- **Validation**: Downloaded schemas/docs must be hash-verified or cryptographically signed.
- **Caching**: Must be aggressively cached locally to prevent rate limiting or external dependency during runtime.
- **Frequency**: Must occur out-of-band (e.g., during a specific `kubeguard update` command), never dynamically during a scan.
- **Provenance**: Must track the source URL, timestamp, and version of the collected data.
- **Failure Behavior**: Must fail gracefully to local cached copies. A network failure must never fail a scan.
- **Licensing**: Ensure scraped data complies with Apache 2.0 or compatible permissive open-source licenses.
