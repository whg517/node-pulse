import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import ProblemDiagnosis from '../ProblemDiagnosis'

describe('ProblemDiagnosis', () => {
  it('renders problem type and confidence', () => {
    render(
      <ProblemDiagnosis problemType="node_local" confidence="high" />
    )

    expect(screen.getByText('Node Local Fault')).toBeInTheDocument()
    expect(screen.getByText(/High Confidence/)).toBeInTheDocument()
  })

  it('renders correct configuration for node_local', () => {
    render(
      <ProblemDiagnosis problemType="node_local" confidence="high" />
    )

    expect(screen.getByText('Node Local Fault')).toBeInTheDocument()
    expect(screen.getByText('Issue detected on this specific node only')).toBeInTheDocument()
    // Chinese text is only visible when expanded
    expect(screen.queryByText('节点本地故障')).not.toBeInTheDocument()
  })

  it('renders correct configuration for cross_border_link', () => {
    render(
      <ProblemDiagnosis problemType="cross_border_link" confidence="medium" />
    )

    expect(screen.getByText('Cross-Border Link Issue')).toBeInTheDocument()
    expect(screen.getByText(/Medium Confidence/)).toBeInTheDocument()
  })

  it('renders correct configuration for carrier_routing', () => {
    render(
      <ProblemDiagnosis problemType="carrier_routing" confidence="low" />
    )

    expect(screen.getByText('Carrier Routing Issue')).toBeInTheDocument()
    expect(screen.getByText(/Low Confidence/)).toBeInTheDocument()
  })

  it('renders correct configuration for none', () => {
    render(
      <ProblemDiagnosis problemType="none" confidence="high" />
    )

    expect(screen.getByText('No Issues Detected')).toBeInTheDocument()
    expect(screen.getByText('All metrics are within normal ranges')).toBeInTheDocument()
  })

  it('renders details when provided', () => {
    render(
      <ProblemDiagnosis
        problemType="node_local"
        confidence="high"
        details="High latency detected on this node only"
      />
    )

    // Details should not be visible initially (collapsed)
    expect(screen.queryByText('High latency detected on this node only')).not.toBeInTheDocument()
  })

  it('expands and collapses on click', () => {
    const { container } = render(
      <ProblemDiagnosis
        problemType="node_local"
        confidence="high"
        details="Test details"
      />
    )

    const clickableDiv = container.querySelector('[role="button"]') as HTMLElement
    expect(clickableDiv).toHaveAttribute('aria-expanded', 'false')

    // Click to expand
    fireEvent.click(clickableDiv)
    expect(clickableDiv).toHaveAttribute('aria-expanded', 'true')
    expect(screen.getByText('Test details')).toBeInTheDocument()

    // Click to collapse
    fireEvent.click(clickableDiv)
    expect(clickableDiv).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByText('Test details')).not.toBeInTheDocument()
  })

  it('expands with keyboard interaction', () => {
    const { container } = render(
      <ProblemDiagnosis
        problemType="node_local"
        confidence="high"
        details="Test details"
      />
    )

    const clickableDiv = container.querySelector('[role="button"]') as HTMLElement

    // Press Enter to expand
    fireEvent.keyDown(clickableDiv, { key: 'Enter', code: 'Enter' })
    expect(clickableDiv).toHaveAttribute('aria-expanded', 'true')

    // Press Space to collapse
    fireEvent.keyDown(clickableDiv, { key: ' ', code: 'Space' })
    expect(clickableDiv).toHaveAttribute('aria-expanded', 'false')
  })

  it('renders in expanded state when isExpanded is true', () => {
    render(
      <ProblemDiagnosis
        problemType="node_local"
        confidence="high"
        details="Test details"
        isExpanded={true}
      />
    )

    expect(screen.getByText('Test details')).toBeInTheDocument()
  })

  it('applies custom className', () => {
    const { container } = render(
      <ProblemDiagnosis
        problemType="node_local"
        confidence="high"
        className="custom-class"
      />
    )

    const card = container.querySelector('.problem-diagnosis')
    expect(card).toHaveClass('custom-class')
  })

  it('has proper ARIA attributes', () => {
    const { container } = render(
      <ProblemDiagnosis problemType="node_local" confidence="high" />
    )

    expect(screen.getByRole('region', { name: 'Problem diagnosis' })).toBeInTheDocument()
    // Find the clickable div with role="button"
    const clickableDiv = container.querySelector('[role="button"]')
    expect(clickableDiv).toHaveAttribute('aria-expanded', 'false')
  })

  it('shows Chinese translation when expanded', () => {
    render(
      <ProblemDiagnosis
        problemType="node_local"
        confidence="high"
        isExpanded={true}
      />
    )

    expect(screen.getByText('节点本地故障')).toBeInTheDocument()
    expect(screen.getByText('仅在此节点检测到问题')).toBeInTheDocument()
  })

  it('shows diagnostic note when expanded', () => {
    render(
      <ProblemDiagnosis
        problemType="node_local"
        confidence="high"
        isExpanded={true}
      />
    )

    expect(screen.getByText(/Note: This is an automated assessment/)).toBeInTheDocument()
  })
})
