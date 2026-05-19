import { Plus } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Alert, Badge, Button, Card, Col, Form, Modal, ProgressBar, Row } from 'react-bootstrap'

import { api, buildQuery } from '../../shared/api.js'
import { money } from '../../shared/format.js'

export function Budgets({ app }) {
  const now = new Date()
  const [month] = useState(now.getMonth() + 1)
  const [year] = useState(now.getFullYear())
  const [budgets, setBudgets] = useState([])
  const [categories, setCategories] = useState([])
  const [showModal, setShowModal] = useState(false)
  const [draft, setDraft] = useState({ category_id: '', label: 'Overall', amount: '', carry_over: false })
  const [message, setMessage] = useState(null)

  useEffect(() => {
    loadBudgets()
    api('/api/v1/categories')
      .then((data) => setCategories(Array.isArray(data) ? data : []))
      .catch(() => setCategories([]))
  }, [])

  async function loadBudgets() {
    try {
      const data = await api(`/api/v1/budgets${buildQuery({ year, month })}`)
      setBudgets(Array.isArray(data) ? data : [])
    } catch (error) {
      setMessage({ variant: 'danger', text: error.message })
    }
  }

  function openBudget(categoryID = '', label = 'Overall') {
    const existing = budgets.find((budget) => (budget.category_id || '') === (categoryID || ''))
    setDraft({
      category_id: categoryID || '',
      label,
      amount: existing?.amount || existing?.effective || '',
      carry_over: Boolean(existing?.carry_over),
    })
    setShowModal(true)
  }

  async function saveBudget(event) {
    event.preventDefault()
    const amount = Number(draft.amount)
    if (!amount || amount <= 0) {
      setMessage({ variant: 'warning', text: 'Enter a valid budget amount.' })
      return
    }
    try {
      await api('/api/v1/budgets', {
        method: 'PUT',
        body: JSON.stringify({
          category_id: draft.category_id || null,
          amount,
          month,
          year,
          carry_over: draft.carry_over,
        }),
      })
      setShowModal(false)
      setMessage({ variant: 'success', text: 'Budget saved.' })
      await loadBudgets()
      app.refresh()
    } catch (error) {
      setMessage({ variant: 'danger', text: error.message })
    }
  }

  const currency = app.settings.currency || 'BDT'

  return (
    <div className="page-stack">
      {message && (
        <Alert variant={message.variant} dismissible onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <div className="section-toolbar">
        <div>
          <div className="section-title">Budgets</div>
          <div className="muted-text">Current month budget limits and alert position.</div>
        </div>
        <Button variant="primary" onClick={() => openBudget()}>
          <Plus size={16} />
          Set overall budget
        </Button>
      </div>

      <Row className="g-3">
        {budgets.length ? (
          budgets.map((budget) => {
            const pct = Math.min(Number(budget.pct || 0), 100)
            const over = Number(budget.pct || 0) >= 100
            const warn = Number(budget.pct || 0) >= 80
            return (
              <Col md={6} xl={4} key={budget.budget_id || budget.category_id || 'overall'}>
                <Card className="app-card h-100 budget-card">
                  <Card.Body>
                    <div className="budget-head">
                      <div>
                        <div className="budget-title">
                          <span>{budget.category_icon || '▦'}</span>
                          {budget.category_name || 'Overall'}
                        </div>
                        <div className="muted-text">
                          {budget.carry_over ? 'Carry-over enabled' : 'Fixed monthly limit'}
                        </div>
                      </div>
                      <Badge bg={over ? 'danger' : warn ? 'warning' : 'success'}>{Number(budget.pct || 0).toFixed(1)}%</Badge>
                    </div>
                    <ProgressBar now={pct} variant={over ? 'danger' : warn ? 'warning' : 'success'} className="budget-progress" />
                    <div className="budget-meta">
                      <span>{money(budget.spent, currency)} spent</span>
                      <strong>{money(budget.effective, currency)}</strong>
                    </div>
                    <Button variant="outline-secondary" className="w-100 mt-3" onClick={() => openBudget(budget.category_id, budget.category_name || 'Overall')}>
                      Edit budget
                    </Button>
                  </Card.Body>
                </Card>
              </Col>
            )
          })
        ) : (
          <Col>
            <Card className="app-card empty-card">
              <Card.Body>No budgets set yet.</Card.Body>
            </Card>
          </Col>
        )}
      </Row>

      <Card className="app-card">
        <Card.Header>
          <div>
            <Card.Title>Set category budget</Card.Title>
            <Card.Text>Pick a category when you want a specific limit.</Card.Text>
          </div>
        </Card.Header>
        <Card.Body className="category-pills">
          <Button variant="outline-primary" onClick={() => openBudget('', 'Overall')}>Overall</Button>
          {categories.map((category) => (
            <Button key={category.category_id} variant="outline-secondary" onClick={() => openBudget(category.category_id, category.name)}>
              {category.icon} {category.name}
            </Button>
          ))}
        </Card.Body>
      </Card>

      <Modal show={showModal} onHide={() => setShowModal(false)} centered>
        <Form onSubmit={saveBudget}>
          <Modal.Header closeButton>
            <Modal.Title>{draft.label}</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <Form.Label>Budget amount</Form.Label>
            <Form.Control type="number" min="0.01" step="0.01" value={draft.amount} onChange={(event) => setDraft((current) => ({ ...current, amount: event.target.value }))} autoFocus />
            <Form.Check
              className="mt-3"
              type="switch"
              label="Carry unused amount into next month"
              checked={draft.carry_over}
              onChange={(event) => setDraft((current) => ({ ...current, carry_over: event.target.checked }))}
            />
          </Modal.Body>
          <Modal.Footer>
            <Button variant="outline-secondary" onClick={() => setShowModal(false)}>Cancel</Button>
            <Button type="submit" variant="primary">Save</Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  )
}
