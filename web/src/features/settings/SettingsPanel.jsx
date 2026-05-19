import { useState } from 'react'
import { Alert, Card, Col, Form, Row } from 'react-bootstrap'

import { api } from '../../shared/api.js'
import { currencies } from '../../shared/options.js'

export function SettingsPanel({ app }) {
  const [message, setMessage] = useState(null)

  async function saveSetting(key, value) {
    app.setSettings((current) => ({ ...current, [key]: String(value) }))
    try {
      await api('/api/v1/settings', { method: 'PUT', body: JSON.stringify({ [key]: String(value) }) })
      setMessage({ variant: 'success', text: 'Setting saved.' })
      app.refresh()
    } catch (error) {
      setMessage({ variant: 'danger', text: error.message })
    }
  }

  const enabled = app.settings.budget_alert_enabled !== 'false'

  return (
    <div className="page-stack">
      {message && (
        <Alert variant={message.variant} dismissible onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <Card className="app-card">
        <Card.Header>
          <div>
            <Card.Title>Settings</Card.Title>
            <Card.Text>Defaults used by reports, chat, and alerts.</Card.Text>
          </div>
        </Card.Header>
        <Card.Body>
          <Row className="g-4">
            <Col lg={5}>
              <Form.Label>Currency</Form.Label>
              <Form.Select value={app.settings.currency || 'BDT'} onChange={(event) => saveSetting('currency', event.target.value)}>
                {currencies.map((currency) => <option key={currency}>{currency}</option>)}
              </Form.Select>
            </Col>
            <Col lg={7}>
              <Form.Label>Budget alert threshold</Form.Label>
              <div className="range-row">
                <Form.Range
                  min={50}
                  max={100}
                  step={5}
                  value={app.settings.budget_alert_threshold || '80'}
                  onChange={(event) => app.setSettings((current) => ({ ...current, budget_alert_threshold: event.target.value }))}
                  onMouseUp={(event) => saveSetting('budget_alert_threshold', event.currentTarget.value)}
                  onTouchEnd={(event) => saveSetting('budget_alert_threshold', event.currentTarget.value)}
                />
                <strong>{app.settings.budget_alert_threshold || '80'}%</strong>
              </div>
            </Col>
            <Col lg={5}>
              <Form.Check
                type="switch"
                id="budget-alert-toggle"
                label="Telegram budget alerts"
                checked={enabled}
                onChange={(event) => saveSetting('budget_alert_enabled', event.target.checked)}
              />
            </Col>
          </Row>
        </Card.Body>
      </Card>
    </div>
  )
}
