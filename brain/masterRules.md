# Master Rules for KubeGuard

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

These are the highest-level engineering rules for KubeGuard. All contributors must strictly adhere to these rules:

1. **Never fabricate results**: Security scanning and policy evaluation must reflect the absolute truth of the input data.
2. **Never bypass tests**: All code must pass the test suite before merging.
3. **Never weaken security merely to make tests pass**: Tests must be fixed to accommodate strict security standards, not the other way around.
4. **Never give the admission controller unnecessary privileges**: RBAC must follow the principle of least privilege.
5. **Never duplicate policy logic**: Use the shared policy engine across CLI, CI, and Admission interfaces.
6. **Never hardcode credentials**: No API keys, tokens, or secrets in source code.
7. **Never commit kubeconfig**: Ensure `.kube/config` and similar files are excluded via `.gitignore`.
8. **Never commit cloud credentials**: Exclude AWS/GCP/Azure credentials from the repository.
9. **Never use AI as the authoritative security decision-maker**: Core policy evaluations must be deterministic.
10. **Keep CLI and admission evaluation deterministic**: Given the same input and policies, the output must always be identical.
11. **Preserve backward compatibility once public APIs stabilize**: Treat CLI flags, JSON output, and AdmissionReview as stable contracts.
12. **Prefer explainable rules**: Every rule must have a clear `why` and `remediation` field.
13. **Keep dependencies minimal**: Avoid bloat. Audit third-party packages strictly.
14. **Document architectural decisions**: Use ADRs for all major design choices.
15. **Every production feature requires tests**: No code goes to production without unit/integration tests.
16. **Every security rule requires positive and negative test cases**: Ensure it catches bad configs and ignores good ones.
17. **Every failure mode must have documented behavior**: Fail-open vs fail-closed states must be clearly defined.
18. **Every release must pass the production checklist**: No cutting corners on the release process.
19. **Documentation must match implementation**: Outdated docs are bugs.
20. **Never claim Kubernetes compatibility that was not tested**: Mark versions as "NOT YET TESTED" until validated.
