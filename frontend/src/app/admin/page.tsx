"use client";

import {
  ArrowLeft,
  ArrowRight,
  Boxes,
  Check,
  CircleAlert,
  ClipboardList,
  Eye,
  EyeOff,
  LoaderCircle,
  LogOut,
  PackageCheck,
  RefreshCw,
  ShieldCheck,
  Store,
} from "lucide-react";
import { FormEvent, useCallback, useEffect, useState } from "react";
import {
  apiRequest,
  Branch,
  formatMoney,
  InventoryItem,
  Order,
  User,
} from "@/lib/api";

type AdminSession = { token: string; user: User };
type LoginResponse = { token: string; user: User };
type TradeIn = {
  id: string;
  customer_name: string;
  customer_phone: string;
  current_brand: string;
  current_model: string;
  current_year: number;
  customer_expected_price: number;
  status: string;
};
type Section = "overview" | "orders" | "inventory" | "trade-in";

const sections: { id: Section; label: string; icon: typeof Store }[] = [
  { id: "overview", label: "Overview", icon: Store },
  { id: "orders", label: "Orders", icon: ClipboardList },
  { id: "inventory", label: "Inventory", icon: Boxes },
  { id: "trade-in", label: "Trade-ins", icon: PackageCheck },
];

function initials(user: User) {
  return (user.full_name || user.username)
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase())
    .join("");
}

export default function AdminPage() {
  const [session, setSession] = useState<AdminSession | null>(null);
  const [checkingSession, setCheckingSession] = useState(true);
  const [section, setSection] = useState<Section>("overview");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [loggingIn, setLoggingIn] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [orders, setOrders] = useState<Order[]>([]);
  const [inventory, setInventory] = useState<InventoryItem[]>([]);
  const [tradeIns, setTradeIns] = useState<TradeIn[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [loadingData, setLoadingData] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    try {
      const saved = window.localStorage.getItem("vehiclehub-admin-session");
      if (saved) setSession(JSON.parse(saved) as AdminSession);
    } catch {
      window.localStorage.removeItem("vehiclehub-admin-session");
    } finally {
      setCheckingSession(false);
    }
  }, []);

  const loadData = useCallback(async (showLoading = true) => {
    if (!session) return;
    if (showLoading) {
      setLoadingData(true);
      setError("");
    }
    try {
      const responses = await Promise.all([
        apiRequest<{ data: Branch[] }>("/branches"),
        ["SUPER_ADMIN", "BRANCH_ADMIN"].includes(session.user.role)
          ? apiRequest<{ data: Order[] }>("/admin/orders", { token: session.token })
          : Promise.resolve({ data: [] as Order[] }),
        apiRequest<{ data: InventoryItem[] }>("/admin/inventory", { token: session.token }),
        apiRequest<{ data: TradeIn[] }>("/admin/trade-in", { token: session.token }),
      ]);
      setBranches(responses[0].data ?? []);
      setOrders(responses[1].data ?? []);
      setInventory(responses[2].data ?? []);
      setTradeIns(responses[3].data ?? []);
    } catch (loadError) {
      const message =
        loadError instanceof Error ? loadError.message : "Could not load dashboard data.";
      setError(
        message.toLowerCase().includes("token") || message.includes("401")
          ? "Your session has expired. Please sign in again."
          : message,
      );
      if (message.toLowerCase().includes("token") || message.includes("401")) logout();
    } finally {
      setLoadingData(false);
    }
  }, [session]);

  useEffect(() => {
    void loadData();
    const interval = window.setInterval(() => {
      void loadData(false);
    }, 15_000);
    return () => window.clearInterval(interval);
  }, [loadData, refreshKey]);

  function logout() {
    window.localStorage.removeItem("vehiclehub-admin-session");
    setSession(null);
    setOrders([]);
    setInventory([]);
    setTradeIns([]);
  }

  async function submitLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoggingIn(true);
    setError("");
    try {
      const result = await apiRequest<LoginResponse>("/auth/login", {
        method: "POST",
        body: JSON.stringify({ username: username.trim(), password }),
      });
      const nextSession = { token: result.token, user: result.user };
      window.localStorage.setItem("vehiclehub-admin-session", JSON.stringify(nextSession));
      setSession(nextSession);
      setPassword("");
    } catch (loginError) {
      const message =
        loginError instanceof Error ? loginError.message : "Sign in failed.";
      setError(
        message.toLowerCase() === "invalid credentials"
          ? "That email or password doesn't match an active staff account. For the demo, use superadmin with password admin123. Customer emails must first be added as staff accounts."
          : message,
      );
    } finally {
      setLoggingIn(false);
    }
  }

  function fillDemoCredentials() {
    setUsername("superadmin");
    setPassword("admin123");
    setError("");
  }

  async function updateOrderStatus(order: Order, status: string) {
    if (!session) return;
    setError("");
    setNotice("");
    try {
      await apiRequest(`/admin/orders/${encodeURIComponent(order.id)}/status`, {
        method: "PATCH",
        token: session.token,
        body: JSON.stringify({ status }),
      });
      setNotice(`Order ${order.order_no} updated.`);
      setRefreshKey((key) => key + 1);
    } catch (updateError) {
      setError(updateError instanceof Error ? updateError.message : "Could not update order.");
    }
  }

  async function confirmDeliveryFee(order: Order, fee: number) {
    if (!session) return;
    setError("");
    setNotice("");
    try {
      await apiRequest(`/admin/orders/${encodeURIComponent(order.id)}/delivery-fee`, {
        method: "PATCH",
        token: session.token,
        body: JSON.stringify({ delivery_fee: fee }),
      });
      setNotice(`Delivery fee confirmed for ${order.order_no}.`);
      setRefreshKey((key) => key + 1);
    } catch (updateError) {
      setError(updateError instanceof Error ? updateError.message : "Could not update delivery fee.");
    }
  }

  if (checkingSession) {
    return <main className="admin-loading"><LoaderCircle className="spin" /> Preparing your workspace...</main>;
  }

  if (!session) {
    return (
      <main className="admin-login-page">
        <a className="brand-lockup" href="/"><span className="brand-mark"><Store size={20} /></span><span>vehicle<span>hub</span><small>DEALER PORTAL</small></span></a>
        <div className="login-layout">
          <section className="login-story">
            <div className="eyebrow eyebrow-light"><span /> YOUR TEAM&apos;S WORKSPACE</div>
            <h1>Good work<br />starts <em>here.</em></h1>
            <p>A clearer view of your orders, inventory and the people coming in for their next ride.</p>
            <div className="login-story-note"><ShieldCheck size={18} /><span>Secure access for authorized branch teams.</span></div>
            <div className="login-orbit login-orbit-a" /><div className="login-orbit login-orbit-b" />
          </section>
          <section className="login-card">
            <a className="back-link" href="/"><ArrowLeft size={15} /> Back to showroom</a>
            <div className="login-card-heading">
              <span className="login-icon"><ShieldCheck size={21} /></span>
              <div className="eyebrow"><span /> DEALER PORTAL</div>
              <h2>Welcome back.</h2>
              <p>Sign in with your staff account to continue.</p>
            </div>
            {error && <div className="alert alert-error"><CircleAlert size={17} />{error}</div>}
            <form className="login-form" onSubmit={submitLogin}>
              <label>
                Username or email
                <input
                  required
                  autoComplete="username"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  placeholder="e.g. superadmin"
                />
              </label>
              <label>
                Password
                <span className="login-password-field">
                  <input
                    required
                    type={showPassword ? "text" : "password"}
                    autoComplete="current-password"
                    value={password}
                    onChange={(event) => setPassword(event.target.value)}
                    placeholder="Enter your password"
                  />
                  <button
                    className="password-visibility"
                    type="button"
                    aria-label={showPassword ? "Hide password" : "Show password"}
                    aria-pressed={showPassword}
                    onClick={() => setShowPassword((visible) => !visible)}
                  >
                    {showPassword ? <EyeOff size={17} /> : <Eye size={17} />}
                  </button>
                </span>
              </label>
              <button className="button button-dark login-submit" disabled={loggingIn}>
                {loggingIn ? (
                  <><LoaderCircle className="spin" size={17} /> Signing in...</>
                ) : (
                  <>Sign in <ArrowRight size={16} /></>
                )}
              </button>
            </form>
            <div className="login-demo">
              <div>
                <strong>Trying the demo?</strong>
                <span>superadmin <i>·</i> admin123</span>
              </div>
              <button type="button" onClick={fillDemoCredentials}>Use demo account</button>
            </div>
            <p className="login-support">Use your staff username or email. Need access? Ask your system administrator for an account.</p>
          </section>
        </div>
        <footer className="login-footer"><span>© {new Date().getFullYear()} G4 Auto Car</span><span>Dealer workspace</span></footer>
      </main>
    );
  }

  const openOrders = orders.filter(
    (order) => !["COMPLETED", "CANCELLED"].includes(order.status),
  ).length;
  const lowStock = inventory.filter(
    (item) => item.available_qty <= item.min_threshold,
  ).length;
  const canSeeOrders = ["SUPER_ADMIN", "BRANCH_ADMIN"].includes(session.user.role);
  const visibleSections = sections.filter(
    (item) => item.id !== "orders" || canSeeOrders,
  );
  const totalSales = orders.reduce((sum, order) => sum + order.total_amount, 0);

  return (
    <main className="admin-shell">
      <aside className="admin-sidebar">
        <a className="brand-lockup" href="/"><span className="brand-mark"><Store size={20} /></span><span>G4 <span>Auto Car</span><small>DEALER PORTAL</small></span></a>
        <div className="sidebar-label">WORKSPACE</div>
        <nav className="admin-nav" aria-label="Dealer navigation">
          {visibleSections.map((item) => {
            const Icon = item.icon;
            return <button key={item.id} className={section === item.id ? "admin-nav-item active" : "admin-nav-item"} onClick={() => setSection(item.id)}><Icon size={18} />{item.label}{item.id === "orders" && openOrders > 0 && <span className="nav-count">{openOrders}</span>}</button>;
          })}
        </nav>
        <div className="sidebar-bottom">
          <div className="sidebar-help"><CircleAlert size={18} /><strong>Need a hand?</strong><span>Reach out to your system administrator.</span></div>
          <a className="back-link" href="/"><ArrowLeft size={15} /> Return to showroom</a>
        </div>
      </aside>

      <section className="admin-main">
        <header className="admin-topbar"><div><span className="breadcrumb">G4 Auto Car</span><span className="breadcrumb-separator">/</span><strong>{sections.find((item) => item.id === section)?.label}</strong></div><div className="admin-top-actions"><button className="refresh-button" onClick={() => setRefreshKey((key) => key + 1)} disabled={loadingData}><RefreshCw size={16} className={loadingData ? "spin" : ""} /> Refresh</button><div className="admin-user"><span className="user-avatar">{initials(session.user)}</span><span><strong>{session.user.full_name || session.user.username}</strong><small>{session.user.role.replaceAll("_", " ")}</small></span></div><button className="logout-button" onClick={logout} aria-label="Sign out"><LogOut size={17} /></button></div></header>

        <div className="admin-content">
          <div className="admin-welcome"><div><div className="eyebrow"><span /> YOUR BUSINESS AT A GLANCE</div><h1>{section === "overview" ? "Good morning." : sections.find((item) => item.id === section)?.label}</h1><p>Here&apos;s what&apos;s happening across your workspace.</p></div><div className="today-date">{new Intl.DateTimeFormat("en", { weekday: "long", month: "short", day: "numeric" }).format(new Date())}</div></div>

          {error && <div className="alert alert-error admin-alert"><CircleAlert size={17} />{error}<button onClick={() => setRefreshKey((key) => key + 1)}>Try again</button></div>}
          {notice && <div className="alert alert-success admin-alert"><Check size={17} />{notice}<button onClick={() => setNotice("")}>Dismiss</button></div>}

          {section === "overview" && (
            <>
              <div className="metric-grid">
                <article className="metric-card"><span className="metric-icon orange"><ClipboardList size={19} /></span><span className="metric-label">Open orders</span><strong>{openOrders}</strong><small>Orders that need attention</small></article>
                <article className="metric-card"><span className="metric-icon green"><Store size={19} /></span><span className="metric-label">Branches</span><strong>{branches.filter((branch) => branch.is_active).length}</strong><small>Active locations</small></article>
                <article className="metric-card"><span className="metric-icon blue"><Boxes size={19} /></span><span className="metric-label">Low stock</span><strong>{lowStock}</strong><small>Items at or under threshold</small></article>
                <article className="metric-card"><span className="metric-icon lilac"><PackageCheck size={19} /></span><span className="metric-label">Order value</span><strong>{formatMoney(totalSales)}</strong><small>Value across loaded orders</small></article>
              </div>
              <div className="dashboard-grid">
                <section className="data-card">
                  <div className="data-card-heading"><div><h2>Recent orders</h2><p>Keep an eye on the latest activity.</p></div><button className="subtle-link" onClick={() => setSection("orders")}>View all <ArrowRight size={15} /></button></div>
                  {loadingData ? <div className="table-loading"><LoaderCircle className="spin" /> Loading orders...</div> : <OrderTable orders={orders.slice(0, 5)} onStatusChange={updateOrderStatus} onDeliveryFeeConfirm={confirmDeliveryFee} />}
                </section>
                <section className="data-card branch-card">
                  <div className="data-card-heading"><div><h2>Your branches</h2><p>Locations connected to your account.</p></div></div>
                  {branches.filter((branch) => branch.is_active).map((branch) => <div className="branch-summary" key={branch.id}><span className="branch-pin"><Store size={16} /></span><span><strong>{branch.name}</strong><small>{branch.city} · {branch.phone}</small></span><span className="status-pill status-active">Active</span></div>)}
                  {!branches.length && !loadingData && <p className="muted-copy">No branch information available.</p>}
                </section>
              </div>
            </>
          )}

          {section === "orders" && <section className="data-card full-data-card"><div className="data-card-heading"><div><h2>Order management</h2><p>Review order progress and confirm location-based delivery fees.</p></div></div>{loadingData ? <div className="table-loading"><LoaderCircle className="spin" /> Loading orders...</div> : <OrderTable orders={orders} onStatusChange={updateOrderStatus} onDeliveryFeeConfirm={confirmDeliveryFee} />}</section>}

          {section === "inventory" && <section className="data-card full-data-card"><div className="data-card-heading"><div><h2>Branch inventory</h2><p>Available stock and reorder thresholds.</p></div></div>{loadingData ? <div className="table-loading"><LoaderCircle className="spin" /> Loading inventory...</div> : inventory.length ? <div className="table-scroll"><table><thead><tr><th>Product</th><th>Branch</th><th>On hand</th><th>Reserved</th><th>Available</th><th>Price</th><th>Stock health</th></tr></thead><tbody>{inventory.map((item) => <tr key={item.id}><td><strong>{item.product_name || "Product"}</strong></td><td>{item.branch_name || "—"}</td><td>{item.quantity}</td><td>{item.reserved_qty}</td><td>{item.available_qty}</td><td>{formatMoney(item.selling_price)}</td><td><span className={item.available_qty <= item.min_threshold ? "status-pill status-low" : "status-pill status-active"}>{item.available_qty <= item.min_threshold ? "Reorder" : "In stock"}</span></td></tr>)}</tbody></table></div> : <div className="empty-admin"><Boxes size={24} /><strong>No inventory to show</strong><span>Inventory records will appear here when available.</span></div>}</section>}

          {section === "trade-in" && <section className="data-card full-data-card"><div className="data-card-heading"><div><h2>Trade-in requests</h2><p>Customer vehicles submitted for appraisal.</p></div></div>{loadingData ? <div className="table-loading"><LoaderCircle className="spin" /> Loading requests...</div> : tradeIns.length ? <div className="table-scroll"><table><thead><tr><th>Customer</th><th>Phone</th><th>Vehicle</th><th>Year</th><th>Expected value</th><th>Status</th></tr></thead><tbody>{tradeIns.map((trade) => <tr key={trade.id}><td><strong>{trade.customer_name || "Customer"}</strong></td><td>{trade.customer_phone || "—"}</td><td>{trade.current_brand} {trade.current_model}</td><td>{trade.current_year || "—"}</td><td>{formatMoney(trade.customer_expected_price)}</td><td><span className="status-pill status-pending">{trade.status.replaceAll("_", " ")}</span></td></tr>)}</tbody></table></div> : <div className="empty-admin"><PackageCheck size={24} /><strong>No trade-in requests yet</strong><span>New customer submissions will appear here.</span></div>}</section>}
        </div>
      </section>
    </main>
  );
}

function OrderTable({
  orders,
  onStatusChange,
  onDeliveryFeeConfirm,
}: {
  orders: Order[];
  onStatusChange: (order: Order, status: string) => Promise<void>;
  onDeliveryFeeConfirm: (order: Order, fee: number) => Promise<void>;
}) {
  const [deliveryFees, setDeliveryFees] = useState<Record<string, string>>({});
  if (!orders.length) {
    return <div className="empty-admin"><ClipboardList size={24} /><strong>No orders yet</strong><span>New customer orders will show up here.</span></div>;
  }
  return <div className="table-scroll"><table><thead><tr><th>Order</th><th>Customer</th><th>Delivery</th><th>Date</th><th>Total</th><th>Status</th><th>Update</th></tr></thead><tbody>{orders.map((order) => <tr key={order.id}><td><strong>{order.order_no}</strong></td><td>{order.customer_name || "Customer"}<small className="table-subline">{order.customer_phone}</small></td><td><strong>{order.delivery_method === "DELIVERY" ? "Store delivery" : "Branch pickup"}</strong>{order.delivery_method === "DELIVERY" && <><small className="table-subline">{order.delivery_address || "Address not provided"}</small>{order.delivery_fee_status === "TO_CONFIRM" ? <form className="delivery-fee-form" onSubmit={(event) => { event.preventDefault(); const fee = Number(deliveryFees[order.id] ?? 0); if (Number.isFinite(fee) && fee >= 0) void onDeliveryFeeConfirm(order, fee); }}><input aria-label={`Delivery fee for ${order.order_no}`} type="number" min="0" step="0.01" value={deliveryFees[order.id] ?? ""} onChange={(event) => setDeliveryFees((current) => ({ ...current, [order.id]: event.target.value }))} placeholder="Fee" required /><button type="submit">Confirm fee</button></form> : <small className="table-subline">Fee: {formatMoney(order.delivery_fee)} · confirmed</small>}</>}</td><td>{new Date(order.created_at).toLocaleDateString()}</td><td>{formatMoney(order.total_amount)}</td><td><span className={`status-pill status-${order.status.toLowerCase()}`}>{order.status.replaceAll("_", " ")}</span></td><td>{["COMPLETED", "CANCELLED"].includes(order.status) ? <span className="muted-copy">—</span> : <select className="status-select" aria-label={`Update order ${order.order_no}`} value={order.status} onChange={(event) => void onStatusChange(order, event.target.value)}><option value={order.status}>{order.status.replaceAll("_", " ")}</option>{["CONFIRMED", "PROCESSING", "READY_FOR_PICKUP", "COMPLETED", "CANCELLED"].filter((status) => status !== order.status).map((status) => <option value={status} key={status}>{status.replaceAll("_", " ")}</option>)}</select>}</td></tr>)}</tbody></table></div>;
}
