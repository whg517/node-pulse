---
title: 'Auto-Sprint Workflow Validation Checklist'
validation-target: 'Auto-sprint orchestration session'
validation-criticality: 'HIGH'
required-inputs:
  - 'Valid sprint-status.yaml file at {sprint_status}'
  - 'Workflow engine available at {project-root}/_bmad/core/tasks/workflow.xml'
  - 'Sub-workflows installed (dev-story, code-review, create-story, retrospective)'
optional-inputs:
  - 'Project context at {project_context}'
  - 'User-provided story file path'
  - 'Custom orchestration settings (max_iterations, skip_epics, etc.)'
validation-rules:
  - 'sprint-status.yaml must exist and be well-formed'
  - 'All status values must be valid for their type (story/epic/retrospective)'
  - 'Sub-workflows must be properly installed and accessible'
  - 'Orchestration loop must have termination conditions'
---

# 🎯 Auto-Sprint Validation Checklist

**Critical validation:** Auto-sprint can only begin when pre-flight checks pass completely

## 📋 Pre-Flight Validation

### Environment Checks

- [ ] **Workflow Engine Available:** Core workflow.xml exists at {project-root}/_bmad/core/tasks/workflow.xml
- [ ] **Sprint Status Exists:** sprint-status.yaml found at {sprint_status}
- [ ] **Config Valid:** {config_source} resolves and contains required fields
- [ ] **Output Directory Writable:** {implementation_artifacts} directory is accessible

### Sub-Workflow Availability

- [ ] **dev-story Installed:** Workflow exists at 4-implementation/dev-story/workflow.yaml
- [ ] **code-review Installed:** Workflow exists at 4-implementation/code-review/workflow.yaml
- [ ] **create-story Installed:** Workflow exists at 4-implementation/create-story/workflow.yaml
- [ ] **retrospective Installed:** Workflow exists at 4-implementation/retrospective/workflow.yaml (optional but recommended)

### Sprint Status Validation

- [ ] **Valid YAML Structure:** sprint-status.yaml parses correctly
- [ ] **Metadata Complete:** generated, project, project_key, tracking_system, story_location fields present
- [ ] **Development Status Section:** development_status map exists with at least one entry
- [ ] **Valid Status Values:** All status values are recognized for their type
  - Stories: backlog, ready-for-dev, in-progress, review, done
  - Epics: backlog, in-progress, done
  - Retrospectives: optional, done
- [ ] **No Orphaned Entries:** All stories belong to valid epics
- [ ] **No Empty Epics:** Epics marked in-progress have at least one associated story

### Orchestration Configuration

- [ ] **Mode Set:** {{mode}} is one of: continuous, single-story, manual
- [ ] **Iteration Limit:** {{max_iterations}} >= 0 (0 = unlimited)
- [ ] **Pause Configuration:** {{pause_on_review}} and {{pause_on_retrospective}} are boolean
- [ ] **Skip List Valid:** {{skip_epics}} contains valid epic keys if non-empty
- [ ] **Termination Conditions:** Clear halt conditions defined (user intervention, iteration limit, completion)

## 🔄 Runtime Validation

### Story Routing Logic

- [ ] **Priority Order Correct:** Searches in-progress → review → ready-for-dev → backlog
- [ ] **Skip List Respected:** Stories from {{skip_epics}} are not selected
- [ ] **Story File Mapped:** Selected story_key maps to existing .md file in {story_dir}
- [ ] **Status Transitions Valid:** Workflow only transitions to valid next states

### Workflow Invocation

- [ ] **Correct Workflow Selected:** Workflow matches story status (create/dev/review)
- [ ] **Parameters Passed:** Story file path and other required parameters provided
- [ ] **Context Preserved:** Communication language, skill level, and other settings maintained
- [ ] **State Tracked:** Previous story and iteration count updated after each workflow

### Error Handling

- [ ] **Missing Workflow Detected:** Halts with clear error if sub-workflow not found
- [ ] **Missing Story File Detected:** Offers recovery options (remove/halt)
- [ ] **Orphaned Story Detected:** Flags stories without valid epic association
- [ ] **Infinite Loop Prevention:** Tracks retry count, halts on threshold
- [ ] **User Interruption Handled:** Completes current task, saves state before halting

## 📊 Progress Tracking

### Status Display

- [ ] **Iteration Header Shown:** Current iteration number displayed at start of each loop
- [ ] **Progress Metrics Shown:** Story counts (backlog, ready, in-progress, review, done) displayed
- [ ] **Current Action Shown:** Target workflow and story clearly communicated
- [ ] **Previous Activity Shown:** Most recent completed story displayed

### Epic Completion Detection

- [ ] **Epic Stories Checked:** Verifies all stories in epic are 'done' before marking epic complete
- [ ] **Retrospective Offered:** Prompts for retrospective when epic completes (if configured)
- [ ] **Epic Status Updated:** Epic marked 'done' after retrospective (or skipped if user chooses)
- [ ] **Multiple Epics Handled:** Correctly handles multiple epics completing in sequence

### Session Completion

- [ ] **Iteration Limit Respected:** Halts when {{max_iterations}} reached (if > 0)
- [ ] **All Stories Complete:** Detects when all stories reach 'done' status
- [ ] **All Epics Complete:** Detects when all epics reach 'done' status
- [ ] **Final Summary Provided:** Displays comprehensive session summary on completion

## 🎯 Validation Output

```
Pre-Flight Validation: {{PASS/FAIL}}

✅ **Auto-Sprint Ready to Execute:**
📊 **Sprint Status:** {{count_done}}/{{total_stories}} stories complete
🔄 **Next Action:** {{next_workflow}} ({{next_story}})
⚙️ **Configuration:** Mode={{mode}}, MaxIterations={{max_iterations}}

{{if FAIL}}
❌ **Validation Failures:**
{{#each failures}}
- {{this}}
{{/each}}

**Required Actions:** {{required_actions}}
{{/if}}
```

## 🚨 Common Validation Failures

### sprint-status.yaml Issues

| Failure | Cause | Resolution |
|---------|-------|------------|
| File not found | sprint-status.yaml doesn't exist | Run `sprint-planning` workflow |
| Invalid YAML | Syntax errors in file | Fix YAML syntax or regenerate |
| Unknown status | Status value not recognized | Correct status value to valid option |
| Missing metadata | Required fields absent | Add metadata or regenerate file |

### Sub-Workflow Issues

| Failure | Cause | Resolution |
|---------|-------|------------|
| Workflow not found | Sub-workflow file missing | Reinstall BMAD module |
| Instructions missing | instructions.xml not found | Reinstall BMAD module |
| Invalid config | workflow.yaml malformed | Reinstall BMAD module |

### Runtime Issues

| Failure | Cause | Resolution |
|---------|-------|------------|
| Story file not found | Story file missing or wrong path | Check story_dir or remove from sprint-status |
| Orphaned story | Story doesn't belong to valid epic | Remove story or fix epic association |
| Infinite loop | Same story retries without progress | Halt and investigate story/workflow |
| Empty epic | Epic has no associated stories | Run correct-course or skip epic |

## ✅ Post-Execution Validation

After auto-sprint completes (or halts):

- [ ] **Session Summary Displayed:** Shows iterations completed, final status, recent activity
- [ ] **State Preserved:** sprint-status.yaml reflects all completed work
- [ ] **Recovery Possible:** User can resume auto-sprint without loss of progress
- [ ] **Next Steps Provided:** Clear guidance on what to do next
- [ ] **No Resource Leaks:** No zombie processes, open files, or locked resources
