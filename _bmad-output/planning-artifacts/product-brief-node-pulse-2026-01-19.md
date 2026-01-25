---
stepsCompleted: [1, 2, 3, 4, 5]
inputDocuments: ["docs/PRD.md"]
date: 2026-01-19
author: Kevin
project_name: node-pulse
---

# Product Brief: NodePulse

<!-- Content will be appended sequentially through collaborative workflow steps -->

## Executive Summary

NodePulse is a distributed network monitoring system designed specifically for overseas nodes connectivity and quality assessment. Consisting of a central management system (Pulse) and lightweight edge detection nodes (Pulse Beacon), system addresses the critical gap in overseas network operations where teams lack effective monitoring tools. By deploying flexible edge nodes that actively monitor network quality through multiple detection methods—ICMP route tracing, business interface detection, and more—NodePulse provides objective data that empowers teams to make informed decisions about cloud service provider selection, edge node optimization, and network troubleshooting.

---

## Core Vision

### Problem Statement

Overseas operations teams lack effective network monitoring tools. Currently, teams can only temporarily deploy cloud hosts for manual detection using ping, traceroute, or simple HTTP scripts, viewing results through console output and making subjective judgments. This approach cannot retain historical objective data, cannot compare and analyze detection results from different locations, and definitely cannot compare them with historical results.

When network issues occur, teams cannot determine the root cause (local node vs. cross-border link vs. operator issues). Operations teams lack data support to make critical decisions—such as whether to switch cloud service providers or add edge nodes in specific regions. This leads to poor access experience for overseas customers and potential business loss.

### Problem Impact

- **Operations Frustration**: Teams respond reactively, with inefficient troubleshooting and no data support for decision-making
- **Business Damage**: Overseas customers experience poor service quality, leading to customer churn
- **Resource Waste**: Each investigation requires temporary cloud host deployment with no reusability
- **Blind Decision-Making**: Unable to assess network quality based on objective data, lacking basis for optimization strategies

### Why Existing Solutions Fall Short

General-purpose network monitoring tools (such as Zabbix, Prometheus, etc.) have the following limitations in overseas node scenarios:

- Lack specialized analysis capabilities for cross-border links
- Cannot easily deploy detection nodes in arbitrary overseas locations
- Lack multi-perspective comparison (different regions, different operators, different time periods)
- More focused on infrastructure monitoring rather than network quality assessment

### Proposed Solution

Build a distributed network monitoring system consisting of a central management system (Pulse) and edge detection nodes (Pulse Beacon):

- **Pulse Beacon**: Deployed in any location requiring monitoring, actively collects network quality data from target nodes through multiple detection methods (ICMP route tracing, business interface detection, etc.)
- **Pulse**: Receives and analyzes detection data reported by Beacons, provides visualization and comparison, alert notifications, and supports integration with existing Prometheus Alertmanager

### Key Differentiators

- **Overseas Scenario Deep Adaptation**: Designed specifically for cross-border network characteristics, supports multi-protocol detection for ICMP-disabled environments
- **Flexible Lightweight Deployment**: Beacon can be deployed in any location with low resource footprint
- **Multi-Perspective Comparative Analysis**: Network quality comparison across different regions, operators, and time periods for objective data-driven decision-making
- **Seamless Integration with Existing Systems**: Supports integration with Prometheus Alertmanager to integrate with existing alerting infrastructure
- **Fast Root Cause Localization**: Distinguishes between local node, cross-border link, and operator issues

---

## Target Users

### Primary Users

#### Operation Supervisor

**Context:**

The operation supervisor manages a network of overseas deployment nodes spread across different regions and cloud providers. They are responsible for ensuring optimal network quality and connectivity for international customers.

**Problem Experience:**

- Currently struggles to identify which overseas nodes have poor network quality in an objective, data-driven way
- Relies on manual, temporary deployments of cloud hosts for ad-hoc testing using ping/traceroute
- Cannot compare network quality across different regions, time periods, or service providers
- Lacks historical data to track trends or validate optimization decisions
- Makes subjective judgments about network quality without evidence
- Faces pressure from overseas customers reporting poor service experiences

**Success Vision:**

- Real-time visibility into which overseas nodes are experiencing network issues, with objective metrics
- Ability to compare network quality across different regions, cloud providers, and time periods
- Historical data and trend analysis to support decision-making
- Clear data to justify decisions like:
  - Replacing poorly performing overseas nodes
  - Switching to a different cloud service provider
  - Adding new edge node deployments in specific regions
- Proactive alerting before network issues impact customers

**"Aha!" Moment:**

When they first see a comprehensive dashboard showing red/yellow/green status across all overseas nodes, with drill-down capability to view root cause (local node vs. cross-border link vs. operator issue), enabling confident, data-driven decisions.

### User Journey

**Discovery:**

Operation supervisor discovers NodePulse through research on overseas network monitoring tools, or hears about it from peers in the operations community.

**Onboarding:**

- Deploys Pulse Beacon to existing overseas nodes (one-click installation)
- Sets up Pulse central management system (Docker deployment)
- Registers Beacons with Pulse, which auto-discovers and starts collecting network metrics
- Configures initial detection targets and thresholds

**Core Usage:**

- Daily: Checks real-time dashboard to see overall network health status across all overseas nodes
- When alerts trigger: Investigates specific nodes showing red/yellow status, reviews metrics and root cause analysis
- Weekly/Monthly: Generates and reviews performance reports, comparing metrics across regions and time periods
- Decision-making: Uses historical data and multi-perspective comparison to make informed decisions about node replacement, provider switching, or edge node additions

**Success Moment:**

When they first use NodePulse to identify a consistently poorly-performing node, pull up comparative data showing it's underperforming relative to similar nodes in other regions, and confidently make a data-backed decision to replace it—saving hours of manual investigation and preventing customer churn.

**Long-term:**

NodePulse becomes integral to their daily operations workflow, providing continuous visibility and enabling proactive network optimization. They build trust in objective metrics and use them consistently for all overseas network decisions.

---

## Success Metrics

### User Success Metrics

**Prompt Problem Detection:**

- Network issues detected and alerts issued within 30 seconds of threshold breach (configurable)
- Operation supervisors receive DingTalk notifications in real-time when problems occur

**Data Visibility:**

- Dashboard loads and displays real-time status of all overseas nodes within 5 seconds
- Historical data viewable for 7-day periods with node label-based comparison
- Operation supervisors can view current status AND historical trends in single interface

**Decision Support:**

- Supervisors use collected data to compare node performance across labels (regions, providers, etc.)
- Objective metrics enable confident decisions about node replacement, provider switching, or edge node additions

**"It's Worth It" Moment:**

When supervisor sees a DingTalk alert, opens dashboard, checks historical comparison data showing that node has been degrading for 3 days, and confidently makes a data-backed decision to replace it—all within 5 minutes of the alert.

### Business Objectives

N/A

### Key Performance Indicators

**System Performance KPIs:**

- Alert delivery success rate: ≥95% of threshold breaches trigger DingTalk notifications
- Dashboard load time: ≤5 seconds from login to full node status display
- Data availability: Historical 7-day data accessible 99% of time

**User Adoption KPIs:**

- Daily active users: Operation supervisors log in at least once daily to check status
- Alert response rate: ≥80% of received DingTalk alerts are viewed and acted upon within 1 hour

**Decision Impact KPIs:**

- Data-backed decisions: ≥70% of node replacement/provider switching decisions reference NodePulse data
- Issue resolution time: Average time from alert to decision reduced by 50% vs. previous manual approach

---

## MVP Scope

### Core Features

**Detection Capabilities:**

- Basic TCP/UDP ping detection (other detection methods deferred to future iterations)
- Verify PulseBeacon ↔ NodePulse communication is functioning
- System health monitoring to verify other NodePulse functions are working properly

**Dashboard & Visualization:**

- View node status by detection protocol perspective
- Display list of registered nodes
- Show detailed information for each node
- Real-time health status indicators (red/yellow/green)

**Alerting:**

- Simple webhook support for alert notifications
- Configurable alert response time (e.g., 30 seconds default)

### Out of Scope for MVP

**Detection Methods:** ICMP ping, MTR/traceroute, iperf3, HTTP monitoring, business traffic monitoring

**Alerting Integrations:** Dedicated DingTalk, SMS, email, or Slack integrations (deferred - webhook only for MVP)

**Security & Authentication:** OIDC login, authorization, encryption algorithms for Beacon-Pulse communication, two-way authentication

**Observability:** OTEL (OpenTelemetry) support for both PulseBeacon and NodePulse

**Management Features:** Audit logging for system event viewing, enhanced system operation management

**External Integration:** Alertmanager integration for pushing alerts to existing monitoring infrastructure

### MVP Success Criteria

**Deployment & Launch:**

- MVP version completed and successfully deployed
- System running stably with basic TCP/UDP ping detection
- Communication between PulseBeacon and NodePulse verified and working

**Validation for Future Iterations:**

- MVP provides foundation for adding more detection methods (route tracking, bandwidth, HTTP requests)
- Establishes architecture for security enhancements (encryption algorithms, two-way authentication)
- Creates platform for observability upgrades (OTEL support)
- Enables system management expansion (OIDC, audit logging)

### Future Vision

**Phase 1: Enhanced Detection & Security**

- Implement additional detection methods:
  - Route tracking/MTR/traceroute
  - Bandwidth testing
  - HTTP request monitoring
- Communication security:
  - Encryption algorithms for Beacon-Pulse communication
  - Two-way authentication mechanism
  - Secure registration process

**Phase 2: System Management & Observability**

- Implement OIDC login for authentication and authorization
- Enhanced system operation management functions
- Audit logging system for viewing system events
- OTEL observability support for both PulseBeacon and NodePulse

**Phase 3: External Integration**

- Support pushing alerts to Prometheus Alertmanager
- Integrate with existing monitoring infrastructure
- Dedicated alert channel integrations (DingTalk, SMS, email, Slack)

**Long-term Vision:**

NodePulse evolves from a focused ping-detection tool into a comprehensive overseas network monitoring platform with:
- Multi-protocol detection capabilities
- Enterprise-grade security and authentication
- Full observability and audit trail
- Seamless integration with existing Prometheus ecosystem
- Advanced features like route topology visualization and AI-powered anomaly detection

---
