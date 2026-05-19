import { useEffect, useState } from 'react'
import { Alert, Button, Card, Col, Form, Modal, Row } from 'react-bootstrap'

import { api } from '../../shared/api.js'

export function Categories() {
  const [categories, setCategories] = useState([])
  const [editing, setEditing] = useState(null)
  const [message, setMessage] = useState(null)

  useEffect(() => {
    loadCategories()
  }, [])

  async function loadCategories() {
    try {
      const data = await api('/api/v1/categories')
      setCategories(Array.isArray(data) ? data : [])
    } catch (error) {
      setMessage({ variant: 'danger', text: error.message })
    }
  }

  async function saveCategory(event) {
    event.preventDefault()
    if (!editing?.name?.trim()) {
      setMessage({ variant: 'warning', text: 'Category name is required.' })
      return
    }
    try {
      await api(`/api/v1/categories/${editing.category_id}`, {
        method: 'PUT',
        body: JSON.stringify({
          name: editing.name.trim(),
          icon: editing.icon || '',
          color: editing.color || '',
        }),
      })
      setEditing(null)
      setMessage({ variant: 'success', text: 'Category updated.' })
      loadCategories()
    } catch (error) {
      setMessage({ variant: 'danger', text: error.message })
    }
  }

  return (
    <div className="page-stack">
      {message && (
        <Alert variant={message.variant} dismissible onClose={() => setMessage(null)}>
          {message.text}
        </Alert>
      )}

      <div className="section-toolbar">
        <div>
          <div className="section-title">Categories</div>
          <div className="muted-text">Edit display names, icons, and colors used by reports and LLM matching.</div>
        </div>
      </div>

      <Row className="g-3">
        {categories.map((category) => (
          <Col sm={6} lg={4} xl={3} key={category.category_id}>
            <Card className="app-card category-card h-100">
              <Card.Body>
                <div className="category-card-head">
                  <span className="category-emoji">{category.icon || '•'}</span>
                  <Button variant="outline-secondary" size="sm" onClick={() => setEditing(category)}>
                    Edit
                  </Button>
                </div>
                <h3>{category.name}</h3>
                <div className="color-swatch" style={{ backgroundColor: category.color || '#64748b' }} />
              </Card.Body>
            </Card>
          </Col>
        ))}
      </Row>

      <Modal show={Boolean(editing)} onHide={() => setEditing(null)} centered>
        <Form onSubmit={saveCategory}>
          <Modal.Header closeButton>
            <Modal.Title>Edit category</Modal.Title>
          </Modal.Header>
          <Modal.Body>
            <Form.Group className="mb-3">
              <Form.Label>Name</Form.Label>
              <Form.Control value={editing?.name || ''} onChange={(event) => setEditing((current) => ({ ...current, name: event.target.value }))} />
            </Form.Group>
            <Form.Group className="mb-3">
              <Form.Label>Icon</Form.Label>
              <Form.Control value={editing?.icon || ''} onChange={(event) => setEditing((current) => ({ ...current, icon: event.target.value }))} />
            </Form.Group>
            <Form.Group>
              <Form.Label>Color</Form.Label>
              <Form.Control type="color" value={editing?.color || '#2563eb'} onChange={(event) => setEditing((current) => ({ ...current, color: event.target.value }))} />
            </Form.Group>
          </Modal.Body>
          <Modal.Footer>
            <Button variant="outline-secondary" onClick={() => setEditing(null)}>Cancel</Button>
            <Button type="submit" variant="primary">Save</Button>
          </Modal.Footer>
        </Form>
      </Modal>
    </div>
  )
}
