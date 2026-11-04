# Laya shadow agreement

generated_at: 2026-11-04T06:49:15.47271Z
laya_version_expected: 0.4.22
model_pin: english
weights_sha256: 891102d372688fc2a094dac56a384bc537b87c63f21f9f3dac0be2b7cbc8d86c
mode: shadow
engram_log: skipped
overall_pass: false

| path | total | agree | errors | display_pct | threshold_pct | pass |
| --- | --- | --- | --- | --- | --- | --- |
| routine_go_nogo | 20 | 8 | 0 | 40.0 | 90 | false |
| pr_class | 20 | 13 | 0 | 65.0 | 95 | false |
| tool_failure_stop | 12 | 3 | 0 | 25.0 | null | null |
| audit_sort | 16 | 4 | 0 | 25.0 | null | null |

## Disagreements

- id=gng-nogo-name-01 actual=no_go shadow=go err=
- id=gng-nogo-name-02 actual=no_go shadow=go err=
- id=gng-nogo-name-03 actual=no_go shadow=go err=
- id=gng-nogo-name-04 actual=no_go shadow=go err=
- id=gng-nogo-read-01 actual=no_go shadow=go err=
- id=gng-nogo-read-02 actual=no_go shadow=go err=
- id=gng-nogo-read-03 actual=no_go shadow=go err=
- id=gng-nogo-read-04 actual=no_go shadow=go err=
- id=gng-nogo-blocked-01 actual=no_go shadow=go err=
- id=gng-nogo-blocked-02 actual=no_go shadow=go err=
- id=gng-nogo-blocked-03 actual=no_go shadow=go err=
- id=gng-nogo-blocked-04 actual=no_go shadow=go err=
- id=pr-t2-new-module actual=T2 shadow=T3 err=
- id=pr-t2-cross-cutting actual=T2 shadow=T3 err=
- id=pr-t2-existing-and-new-module actual=T2 shadow=T0 err=
- id=pr-t2-cross-and-docs actual=T2 shadow=T0 err=
- id=pr-t2-new-module-b actual=T2 shadow=T3 err=
- id=pr-t0-config actual=T0 shadow=T1 err=
- id=pr-t0-config-b actual=T0 shadow=T1 err=
- id=stop-both-count-0 actual=continue shadow=stop err=
- id=stop-both-count-1 actual=continue shadow=stop err=
- id=stop-both-count-2 actual=continue shadow=stop err=
- id=stop-tool-diff-count-3 actual=continue shadow=stop err=
- id=stop-class-diff-count-3 actual=continue shadow=stop err=
- id=stop-both-diff-count-4 actual=continue shadow=stop err=
- id=stop-tool-diff-count-4 actual=continue shadow=stop err=
- id=stop-class-diff-count-5 actual=continue shadow=stop err=
- id=stop-both-diff-count-5 actual=continue shadow=stop err=
- id=audit-security-01 actual=blocker shadow=note err=
- id=audit-security-02 actual=blocker shadow=note err=
- id=audit-security-03 actual=blocker shadow=note err=
- id=audit-data-loss-01 actual=blocker shadow=note err=
- id=audit-data-loss-02 actual=blocker shadow=note err=
- id=audit-data-loss-03 actual=blocker shadow=note err=
- id=audit-behavior-bug-01 actual=should_fix shadow=note err=
- id=audit-behavior-bug-02 actual=should_fix shadow=note err=
- id=audit-behavior-bug-03 actual=should_fix shadow=note err=
- id=audit-naming-01 actual=nit shadow=note err=
- id=audit-naming-02 actual=nit shadow=note err=
- id=audit-naming-03 actual=nit shadow=note err=
