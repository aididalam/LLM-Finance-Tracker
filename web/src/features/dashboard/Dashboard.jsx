import {
  ArcElement,
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from "chart.js";
import { ArrowDownRight, ArrowUpRight, Calendar, WalletCards } from "lucide-react";
import { useEffect, useState } from "react";
import { Badge, Button, Card, Col, ProgressBar, Row, Table } from "react-bootstrap";
import { Doughnut, Line } from "react-chartjs-2";

import { api, buildQuery } from "../../shared/api.js";
import { accountTypeLabel, money, monthLabel } from "../../shared/format.js";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, ArcElement, Tooltip, Legend, Filler);

export function Dashboard({ app }) {
  const now = new Date();
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [overview, setOverview] = useState(null);
  const [trend, setTrend] = useState([]);
  const [wallets, setWallets] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    Promise.all([
      api(`/api/v1/dashboard/overview${buildQuery({ year, month })}`),
      api("/api/v1/dashboard/trend"),
      api("/api/v1/accounts"),
    ])
      .then(([overviewData, trendData, walletData]) => {
        if (cancelled) return;
        setOverview(overviewData);
        setTrend(Array.isArray(trendData) ? trendData : []);
        setWallets(Array.isArray(walletData) ? walletData : []);
      })
      .finally(() => !cancelled && setLoading(false));
    return () => {
      cancelled = true;
    };
  }, [year, month]);

  const currency = overview?.currency || app.settings.currency || "BDT";
  const expense = overview?.this_month || 0;
  const income = overview?.income_this_month || 0;
  const net = overview?.net_this_month || 0;

  function moveMonth(delta) {
    const next = new Date(year, month - 1 + delta, 1);
    setYear(next.getFullYear());
    setMonth(next.getMonth() + 1);
  }

  return (
    <div className="page-stack">
      <div className="section-toolbar">
        <div>
          <div className="section-title">Monthly position</div>
          <div className="muted-text">{monthLabel(year, month)}</div>
        </div>
        <div className="toolbar-actions">
          <Button variant="outline-secondary" onClick={() => moveMonth(-1)}>
            Previous
          </Button>
          <Button variant="outline-secondary" onClick={() => moveMonth(1)}>
            Next
          </Button>
        </div>
      </div>

      <Row className="g-3">
        <Col md={6} xl={3}>
          <StatCard
            label="Income"
            value={money(income, currency)}
            tone="success"
            icon={<ArrowUpRight size={18} />}
          />
        </Col>
        <Col md={6} xl={3}>
          <StatCard
            label="Expenses"
            value={money(expense, currency)}
            tone="danger"
            icon={<ArrowDownRight size={18} />}
          />
        </Col>
        <Col md={6} xl={3}>
          <StatCard
            label="Net"
            value={`${net >= 0 ? "+" : ""}${money(net, currency)}`}
            tone={net >= 0 ? "success" : "danger"}
            icon={<WalletCards size={18} />}
          />
        </Col>
        <Col md={6} xl={3}>
          <StatCard
            label="Last month net"
            value={`${(overview?.net_last_month || 0) >= 0 ? "+" : ""}${money(overview?.net_last_month || 0, currency)}`}
            tone="neutral"
            icon={<Calendar size={18} />}
          />
        </Col>
      </Row>

      <Row className="g-3">
        <Col xl={8}>
          <Card className="app-card h-100">
            <Card.Header>
              <div>
                <Card.Title>12 month trend</Card.Title>
                <Card.Text>Income, expenses, and net movement.</Card.Text>
              </div>
            </Card.Header>
            <Card.Body className="chart-panel">
              {trend.length ? (
                <Line data={trendData(trend)} options={trendOptions(currency)} />
              ) : (
                <EmptyState loading={loading} label="No trend data yet." />
              )}
            </Card.Body>
          </Card>
        </Col>
        <Col xl={4}>
          <Card className="app-card h-100">
            <Card.Header>
              <div>
                <Card.Title>Expense mix</Card.Title>
                <Card.Text>Top categories this month.</Card.Text>
              </div>
            </Card.Header>
            <Card.Body className="chart-panel compact">
              {overview?.categories?.length ? (
                <Doughnut data={donutData(overview.categories)} options={donutOptions(currency)} />
              ) : (
                <EmptyState loading={loading} label="No expenses this month." />
              )}
            </Card.Body>
          </Card>
        </Col>
      </Row>

      <Row className="g-3">
        <Col lg={6}>
          <CategoryList title="Expense categories" items={overview?.categories || []} currency={currency} />
        </Col>
        <Col lg={6}>
          <CategoryList
            title="Income categories"
            items={overview?.income_categories || []}
            currency={currency}
          />
        </Col>
      </Row>

      <Card className="app-card">
        <Card.Header>
          <div>
            <Card.Title>Wallet balances</Card.Title>
            <Card.Text>Current balance for each active wallet.</Card.Text>
          </div>
          <Badge bg="light" text="dark">
            {wallets.length} wallets
          </Badge>
        </Card.Header>
        <div className="table-wrap">
          <Table hover responsive className="app-table mb-0">
            <thead>
              <tr>
                <th>Wallet</th>
                <th>Type</th>
                <th className="text-end">Balance</th>
              </tr>
            </thead>
            <tbody>
              {wallets.length ? (
                wallets.map((w) => (
                  <tr key={w.id}>
                    <td className="fw-semibold">{w.name}</td>
                    <td className="text-muted">{accountTypeLabel(w.account_type)}</td>
                    <td
                      className={`text-end fw-bold ${(w.balance ?? 0) >= 0 ? "text-success" : "text-danger"}`}
                    >
                      {money(w.balance ?? 0, currency)}
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="3" className="empty-cell">
                    {loading ? "Loading wallets..." : "No wallets yet."}
                  </td>
                </tr>
              )}
            </tbody>
          </Table>
        </div>
      </Card>
    </div>
  );
}

function StatCard({ label, value, icon, tone }) {
  return (
    <Card className={`stat-card tone-${tone}`}>
      <Card.Body>
        <div className="stat-head">
          <span>{label}</span>
          <div className="stat-icon">{icon}</div>
        </div>
        <div className="stat-value">{value}</div>
      </Card.Body>
    </Card>
  );
}

function CategoryList({ title, items, currency }) {
  const max = Math.max(...items.map((item) => Number(item.total || 0)), 0);
  return (
    <Card className="app-card h-100">
      <Card.Header>
        <div>
          <Card.Title>{title}</Card.Title>
          <Card.Text>
            {items.length ? `${items.length} active categories` : "Nothing recorded yet."}
          </Card.Text>
        </div>
      </Card.Header>
      <Card.Body className="category-stack">
        {items.length ? (
          items.map((item) => {
            const pct = max ? (Number(item.total || 0) / max) * 100 : 0;
            return (
              <div className="category-row" key={`${title}-${item.category_name}`}>
                <div className="category-icon">{item.icon || "•"}</div>
                <div className="category-body">
                  <div className="category-line">
                    <span>{item.category_name || "Uncategorized"}</span>
                    <strong>{money(item.total, currency)}</strong>
                  </div>
                  <ProgressBar now={pct} className="soft-progress" />
                </div>
                <span className="muted-count">{item.count || 0}x</span>
              </div>
            );
          })
        ) : (
          <EmptyState label="No category data for this month." />
        )}
      </Card.Body>
    </Card>
  );
}

function EmptyState({ loading, label }) {
  return <div className="empty-state">{loading ? "Loading..." : label}</div>;
}

function trendData(points) {
  return {
    labels: points.map((point) => point.month),
    datasets: [
      {
        label: "Income",
        data: points.map((point) => point.income || 0),
        borderColor: "#16a34a",
        backgroundColor: "rgba(22, 163, 74, 0.08)",
        tension: 0.35,
      },
      {
        label: "Expenses",
        data: points.map((point) => point.expense || point.total || 0),
        borderColor: "#e11d48",
        backgroundColor: "rgba(225, 29, 72, 0.08)",
        tension: 0.35,
      },
      {
        label: "Net",
        data: points.map((point) => point.net || 0),
        borderColor: "#2563eb",
        backgroundColor: "rgba(37, 99, 235, 0.12)",
        fill: true,
        tension: 0.35,
      },
    ],
  };
}

function trendOptions(currency) {
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { labels: { boxWidth: 10, usePointStyle: true } },
      tooltip: { callbacks: { label: (ctx) => `${ctx.dataset.label}: ${money(ctx.parsed.y, currency)}` } },
    },
    scales: {
      y: { beginAtZero: true },
    },
  };
}

function donutData(items) {
  return {
    labels: items.map((item) => item.category_name),
    datasets: [
      {
        data: items.map((item) => item.total || 0),
        backgroundColor: [
          "#2563eb",
          "#16a34a",
          "#f59e0b",
          "#e11d48",
          "#7c3aed",
          "#0891b2",
          "#db2777",
          "#65a30d",
        ],
        borderWidth: 0,
      },
    ],
  };
}

function donutOptions(currency) {
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { position: "bottom", labels: { boxWidth: 10, usePointStyle: true } },
      tooltip: { callbacks: { label: (ctx) => `${ctx.label}: ${money(ctx.parsed, currency)}` } },
    },
    cutout: "65%",
  };
}
