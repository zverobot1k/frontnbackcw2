import React, { useEffect, useMemo, useState } from "react";

const API_BASE = import.meta.env.VITE_API_BASE || "/api";
const LOCAL_CART_KEY = "shop_cart_lines";
const ACCESS_KEY = "access_token";
const REFRESH_KEY = "refresh_token";

const emptyAuth = { email: "", password: "", age: 25, gender: "unknown" };
const emptyFilters = { search: "", category: "", minPrice: "", maxPrice: "", inStock: true };
const emptyAdminCreate = { name: "", category: "general", description: "", price: "", stock: "" };
const emptyAdminUpdate = { id: "", name: "", category: "", description: "", price: "", stock: "" };

function readLocalJSON(key, fallback) {
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return fallback;
    return JSON.parse(raw);
  } catch (_err) {
    return fallback;
  }
}

function writeLocalJSON(key, value) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch (_err) {
    // no-op
  }
}

function readToken(key) {
  try {
    return localStorage.getItem(key) || "";
  } catch (_err) {
    return "";
  }
}

function writeToken(key, value) {
  try {
    localStorage.setItem(key, value || "");
  } catch (_err) {
    // no-op
  }
}

function clearToken(key) {
  try {
    localStorage.removeItem(key);
  } catch (_err) {
    // no-op
  }
}

function parseError(payload, status) {
  if (payload && typeof payload === "object" && typeof payload.error === "string") {
    return payload.error;
  }

  return `request failed: ${status}`;
}

function toCurrency(value) {
  return Number(value || 0).toLocaleString("ru-RU", { style: "currency", currency: "USD" });
}

function toDate(value) {
  if (!value) return "-";
  return new Date(value).toLocaleString("ru-RU");
}

function roleTitle(role) {
  if (role === "admin") return "admin";
  if (role === "customer") return "customer";
  return "guest";
}

export default function App() {
  const [busy, setBusy] = useState(false);
  const [status, setStatus] = useState("Готово");
  const [screen, setScreen] = useState("catalog");

  const [authForm, setAuthForm] = useState(emptyAuth);
  const [accessToken, setAccessToken] = useState(() => readToken(ACCESS_KEY));
  const [refreshToken, setRefreshToken] = useState(() => readToken(REFRESH_KEY));
  const [currentUser, setCurrentUser] = useState(null);

  const [filters, setFilters] = useState(emptyFilters);
  const [products, setProducts] = useState([]);

  const [localCart, setLocalCart] = useState(() => readLocalJSON(LOCAL_CART_KEY, []));
  const [cart, setCart] = useState({ items: [], total: 0 });
  const [orders, setOrders] = useState([]);
  const [adminOrders, setAdminOrders] = useState([]);

  const [adminCreate, setAdminCreate] = useState(emptyAdminCreate);
  const [adminUpdate, setAdminUpdate] = useState(emptyAdminUpdate);
  const [deleteProductID, setDeleteProductID] = useState("");

  const isAuth = Boolean(accessToken);
  const role = currentUser?.role || "guest";
  const isAdmin = role === "admin";

  const nav = useMemo(
    () => [
      { key: "catalog", label: "Каталог", enabled: true },
      { key: "cart", label: "Корзина", enabled: true },
      { key: "orders", label: "История", enabled: isAuth },
      { key: "admin", label: "Админ", enabled: isAdmin },
      { key: "auth", label: isAuth ? "Профиль" : "Вход" , enabled: true },
    ],
    [isAuth, isAdmin]
  );

  function authHeaders() {
    return accessToken ? { Authorization: `Bearer ${accessToken}` } : {};
  }

  async function callApi(path, options = {}) {
    const response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
    });

    let payload = null;
    try {
      payload = await response.json();
    } catch (_err) {
      payload = null;
    }

    if (!response.ok) {
      throw new Error(parseError(payload, response.status));
    }

    return payload;
  }

  async function runAction(label, task) {
    setBusy(true);
    setStatus(`${label}...`);

    try {
      await task();
    } catch (err) {
      setStatus(`Ошибка: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  function saveLocalCart(next) {
    setLocalCart(next);
    writeLocalJSON(LOCAL_CART_KEY, next);
  }

  function applyAuth(payload) {
    setAccessToken(payload.access_token || "");
    setRefreshToken(payload.refresh_token || "");
    setCurrentUser(payload.user || null);

    writeToken(ACCESS_KEY, payload.access_token || "");
    writeToken(REFRESH_KEY, payload.refresh_token || "");
  }

  function logout() {
    setAccessToken("");
    setRefreshToken("");
    setCurrentUser(null);
    setCart({ items: [], total: 0 });
    setOrders([]);
    setAdminOrders([]);

    clearToken(ACCESS_KEY);
    clearToken(REFRESH_KEY);
    setScreen("catalog");
    setStatus("Выход выполнен");
  }

  async function registerAndLogin() {
    return runAction("Регистрация", async () => {
      await callApi("/auth/register", {
        method: "POST",
        body: JSON.stringify({
          email: authForm.email,
          password: authForm.password,
          gender: authForm.gender,
          age: Number(authForm.age || 0),
        }),
      });

      const auth = await callApi("/auth/login", {
        method: "POST",
        body: JSON.stringify({
          email: authForm.email,
          password: authForm.password,
        }),
      });

      applyAuth(auth);
      setScreen("catalog");
      setStatus("Регистрация завершена, вы вошли в аккаунт");
    });
  }

  async function login() {
    return runAction("Вход", async () => {
      const auth = await callApi("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email: authForm.email, password: authForm.password }),
      });

      applyAuth(auth);
      setScreen("catalog");
      setStatus("Вход выполнен");
    });
  }

  async function refreshSession() {
    if (!refreshToken) return;

    return runAction("Обновление токена", async () => {
      const auth = await callApi("/auth/refresh", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      applyAuth(auth);
      setStatus("Сессия обновлена");
    });
  }

  async function loadProfile() {
    if (!isAuth) return;

    return runAction("Профиль", async () => {
      const data = await callApi("/auth/me", { headers: authHeaders() });
      setCurrentUser(data);
      setStatus("Профиль обновлен");
    });
  }

  async function loadProducts() {
    return runAction("Каталог", async () => {
      const params = new URLSearchParams();
      if (filters.search.trim()) params.set("search", filters.search.trim());
      if (filters.category.trim()) params.set("category", filters.category.trim());
      if (filters.minPrice.trim()) params.set("min_price", filters.minPrice.trim());
      if (filters.maxPrice.trim()) params.set("max_price", filters.maxPrice.trim());
      if (filters.inStock) params.set("in_stock", "true");

      const suffix = params.toString() ? `?${params.toString()}` : "";
      const data = await callApi(`/products${suffix}`);
      setProducts(Array.isArray(data) ? data : []);
      setStatus("Каталог загружен");
    });
  }

  async function syncCartWithServer(lines = localCart) {
    if (!isAuth) return;

    const payload = { items: lines.map((line) => ({ product_id: line.product_id, quantity: line.quantity })) };

    const serverCart = await callApi("/cart", {
      method: "PUT",
      headers: authHeaders(),
      body: JSON.stringify(payload),
    });

    setCart(serverCart);
    const normalized = (serverCart.items || []).map((line) => ({
      product_id: line.product.id,
      quantity: line.quantity,
    }));
    saveLocalCart(normalized);
  }

  async function loadCart() {
    if (!isAuth) {
      const computedTotal = localCart.reduce((sum, line) => {
        const product = products.find((item) => item.id === line.product_id);
        return sum + Number(product?.price || 0) * line.quantity;
      }, 0);
      setCart({ items: [], total: computedTotal });
      return;
    }

    return runAction("Корзина", async () => {
      const data = await callApi("/cart", { headers: authHeaders() });
      setCart(data);
      const normalized = (data.items || []).map((line) => ({
        product_id: line.product.id,
        quantity: line.quantity,
      }));
      saveLocalCart(normalized);
      setStatus("Корзина обновлена");
    });
  }

  async function addToCart(productID) {
    const found = localCart.find((line) => line.product_id === productID);
    const next = found
      ? localCart.map((line) => (line.product_id === productID ? { ...line, quantity: line.quantity + 1 } : line))
      : [...localCart, { product_id: productID, quantity: 1 }];

    saveLocalCart(next);

    return runAction("Корзина", async () => {
      await syncCartWithServer(next);
      setStatus("Товар добавлен в корзину");
    });
  }

  async function changeLocalQuantity(productID, quantity) {
    const parsed = Number(quantity);
    const next = localCart
      .map((line) => (line.product_id === productID ? { ...line, quantity: parsed } : line))
      .filter((line) => line.quantity > 0);

    saveLocalCart(next);

    return runAction("Синхронизация корзины", async () => {
      await syncCartWithServer(next);
      setStatus("Корзина синхронизирована");
    });
  }

  async function checkout() {
    if (!isAuth) {
      setStatus("Ошибка: для оформления заказа нужна авторизация");
      return;
    }

    return runAction("Оформление заказа", async () => {
      const data = await callApi("/orders/checkout", {
        method: "POST",
        headers: authHeaders(),
      });

      saveLocalCart([]);
      await Promise.all([loadProductsSilent(), loadOrdersSilent(), loadCartSilent()]);

      if (data.stripe_client_secret) {
        setStatus(`Заказ #${data.order_id} создан. Stripe client_secret получен.`);
      } else {
        setStatus(`Заказ #${data.order_id} создан.`);
      }
    });
  }

  async function loadOrders() {
    if (!isAuth) return;

    return runAction("История заказов", async () => {
      const data = await callApi("/orders", { headers: authHeaders() });
      setOrders(Array.isArray(data) ? data : []);
      setStatus("История заказов загружена");
    });
  }

  async function loadAdminOrders() {
    if (!isAdmin) return;

    return runAction("Заказы всех пользователей", async () => {
      const data = await callApi("/admin/orders", { headers: authHeaders() });
      setAdminOrders(Array.isArray(data) ? data : []);
      setStatus("Список заказов для админа обновлен");
    });
  }

  async function createProduct() {
    if (!isAdmin) return;

    return runAction("Создание товара", async () => {
      await callApi("/products", {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({
          name: adminCreate.name,
          category: adminCreate.category,
          description: adminCreate.description,
          price: Number(adminCreate.price || 0),
          stock: Number(adminCreate.stock || 0),
        }),
      });

      setAdminCreate(emptyAdminCreate);
      await loadProductsSilent();
      setStatus("Товар создан");
    });
  }

  async function updateProduct() {
    if (!isAdmin) return;

    return runAction("Обновление товара", async () => {
      const id = Number(adminUpdate.id);
      if (!id) throw new Error("укажите ID товара");

      const payload = {};
      if (adminUpdate.name.trim()) payload.name = adminUpdate.name.trim();
      if (adminUpdate.category.trim()) payload.category = adminUpdate.category.trim();
      if (adminUpdate.description.trim()) payload.description = adminUpdate.description.trim();
      if (adminUpdate.price !== "") payload.price = Number(adminUpdate.price);
      if (adminUpdate.stock !== "") payload.stock = Number(adminUpdate.stock);

      if (Object.keys(payload).length === 0) throw new Error("нет полей для обновления");

      await callApi(`/products/${id}`, {
        method: "PUT",
        headers: authHeaders(),
        body: JSON.stringify(payload),
      });

      await loadProductsSilent();
      setStatus(`Товар #${id} обновлен`);
    });
  }

  async function removeProduct() {
    if (!isAdmin) return;

    return runAction("Удаление товара", async () => {
      const id = Number(deleteProductID);
      if (!id) throw new Error("укажите ID товара");

      await callApi(`/products/${id}`, { method: "DELETE", headers: authHeaders() });
      setDeleteProductID("");
      await loadProductsSilent();
      setStatus(`Товар #${id} удален`);
    });
  }

  async function loadProductsSilent() {
    const params = new URLSearchParams();
    if (filters.search.trim()) params.set("search", filters.search.trim());
    if (filters.category.trim()) params.set("category", filters.category.trim());
    if (filters.minPrice.trim()) params.set("min_price", filters.minPrice.trim());
    if (filters.maxPrice.trim()) params.set("max_price", filters.maxPrice.trim());
    if (filters.inStock) params.set("in_stock", "true");

    const suffix = params.toString() ? `?${params.toString()}` : "";
    const data = await callApi(`/products${suffix}`);
    setProducts(Array.isArray(data) ? data : []);
  }

  async function loadOrdersSilent() {
    if (!isAuth) return;
    const data = await callApi("/orders", { headers: authHeaders() });
    setOrders(Array.isArray(data) ? data : []);
  }

  async function loadCartSilent() {
    if (!isAuth) return;
    const data = await callApi("/cart", { headers: authHeaders() });
    setCart(data);
  }

  useEffect(() => {
    loadProducts();
  }, []);

  useEffect(() => {
    if (!isAuth) return;

    loadProfile();
    syncCartWithServer(localCart);
    loadOrders();
  }, [accessToken]);

  const cartItemsForGuest = localCart.map((line) => {
    const product = products.find((item) => item.id === line.product_id);
    return {
    productID: line.product_id,
      product,
      quantity: line.quantity,
      price: Number(product?.price || 0),
      lineTotal: Number(product?.price || 0) * line.quantity,
    };
  });

  const guestTotal = cartItemsForGuest.reduce((sum, line) => sum + line.lineTotal, 0);

  function renderCatalog() {
    return (
      <section className="panel">
        <div className="panel-head">
          <div>
            <h2>Каталог товаров</h2>
            <p>Фильтрация и поиск по названию, категории и цене.</p>
          </div>
          <button disabled={busy} onClick={loadProducts}>Обновить</button>
        </div>

        <div className="filters">
          <input placeholder="Поиск" value={filters.search} onChange={(e) => setFilters({ ...filters, search: e.target.value })} />
          <input placeholder="Категория" value={filters.category} onChange={(e) => setFilters({ ...filters, category: e.target.value })} />
          <input type="number" placeholder="Мин. цена" value={filters.minPrice} onChange={(e) => setFilters({ ...filters, minPrice: e.target.value })} />
          <input type="number" placeholder="Макс. цена" value={filters.maxPrice} onChange={(e) => setFilters({ ...filters, maxPrice: e.target.value })} />
          <label className="stock-toggle">
            <input type="checkbox" checked={filters.inStock} onChange={(e) => setFilters({ ...filters, inStock: e.target.checked })} />
            Только в наличии
          </label>
          <button disabled={busy} onClick={loadProducts}>Применить</button>
        </div>

        <div className="grid products">
          {products.length === 0 ? <div className="empty">По фильтрам ничего не найдено.</div> : null}
          {products.map((product) => (
            <article className="card product" key={product.id}>
              <div className="row between">
                <h3>{product.name}</h3>
                <span className="pill">#{product.id}</span>
              </div>
              <p className="muted">{product.description}</p>
              <div className="row tags">
                <span className="pill light">{product.category}</span>
                <span className="pill light">Склад: {product.stock}</span>
              </div>
              <div className="row between">
                <strong>{toCurrency(product.price)}</strong>
                <button disabled={busy || product.stock <= 0} onClick={() => addToCart(product.id)}>В корзину</button>
              </div>
            </article>
          ))}
        </div>
      </section>
    );
  }

  function renderCart() {
    const serverItems = isAuth ? cart.items || [] : [];

    return (
      <section className="panel">
        <div className="panel-head">
          <div>
            <h2>Корзина</h2>
            <p>localStorage + синхронизация с сервером после авторизации.</p>
          </div>
          {isAuth ? <button disabled={busy} onClick={loadCart}>Обновить корзину</button> : null}
        </div>

        {isAuth ? (
          <div className="stack">
            {serverItems.length === 0 ? <div className="empty">Корзина пуста.</div> : null}
            {serverItems.map((line) => (
              <div className="line" key={`${line.product.id}_${line.id}`}>
                <div>
                  <strong>{line.product.name}</strong>
                  <p className="muted">{toCurrency(line.price)} x {line.quantity}</p>
                </div>
                <div className="row">
                  <input
                    type="number"
                    min="1"
                    value={line.quantity}
                    onChange={(e) => changeLocalQuantity(line.product.id, e.target.value)}
                  />
                  <span className="pill">{toCurrency(line.price * line.quantity)}</span>
                </div>
              </div>
            ))}
            <div className="row between total">
              <strong>Итого:</strong>
              <strong>{toCurrency(cart.total)}</strong>
            </div>
            <button disabled={busy || serverItems.length === 0} onClick={checkout}>Оформить заказ (Stripe)</button>
          </div>
        ) : (
          <div className="stack">
            {cartItemsForGuest.length === 0 ? <div className="empty">Локальная корзина пуста.</div> : null}
            {cartItemsForGuest.map((line) => (
              <div className="line" key={line.productID}>
                <div>
                <strong>{line.product?.name || `Товар #${line.productID}`}</strong>
                  <p className="muted">{toCurrency(line.price)} x {line.quantity}</p>
                </div>
                <span className="pill">{toCurrency(line.lineTotal)}</span>
              </div>
            ))}
            <div className="row between total">
              <strong>Итого:</strong>
              <strong>{toCurrency(guestTotal)}</strong>
            </div>
            <p className="muted">Войдите в аккаунт, чтобы синхронизировать корзину с сервером и оформить заказ.</p>
          </div>
        )}
      </section>
    );
  }

  function renderOrders() {
    if (!isAuth) {
      return (
        <section className="panel">
          <h2>История заказов</h2>
          <p>Раздел доступен только после авторизации.</p>
        </section>
      );
    }

    return (
      <section className="panel">
        <div className="panel-head">
          <div>
            <h2>История заказов</h2>
            <p>Ваши оформленные заказы.</p>
          </div>
          <button disabled={busy} onClick={loadOrders}>Обновить</button>
        </div>

        <div className="stack">
          {orders.length === 0 ? <div className="empty">Заказов пока нет.</div> : null}
          {orders.map((order) => (
            <article className="card" key={order.id}>
              <div className="row between">
                <h3>Заказ #{order.id}</h3>
                <span className="pill">{order.status}</span>
              </div>
              <p className="muted">Создан: {toDate(order.created_at)}</p>
              <p className="muted">Оплата: {order.payment_provider}, ref: {order.payment_ref || "-"}</p>
              <div className="stack mini">
                {(order.items || []).map((item, idx) => (
                  <div className="line compact" key={`${order.id}_${idx}`}>
                    <span>{item.product_name} x {item.quantity}</span>
                    <strong>{toCurrency(item.line_total)}</strong>
                  </div>
                ))}
              </div>
              <div className="row between total">
                <strong>Итог:</strong>
                <strong>{toCurrency(order.total_amount)}</strong>
              </div>
            </article>
          ))}
        </div>
      </section>
    );
  }

  function renderAdmin() {
    if (!isAdmin) {
      return (
        <section className="panel">
          <h2>Панель администратора</h2>
          <p>Доступ только для роли admin.</p>
        </section>
      );
    }

    return (
      <section className="panel">
        <div className="panel-head">
          <div>
            <h2>Панель администратора</h2>
            <p>Управление товарами и просмотр всех заказов.</p>
          </div>
          <button disabled={busy} onClick={loadAdminOrders}>Загрузить все заказы</button>
        </div>

        <div className="admin-grid">
          <article className="card">
            <h3>Создать товар</h3>
            <input placeholder="Название" value={adminCreate.name} onChange={(e) => setAdminCreate({ ...adminCreate, name: e.target.value })} />
            <input placeholder="Категория" value={adminCreate.category} onChange={(e) => setAdminCreate({ ...adminCreate, category: e.target.value })} />
            <input placeholder="Описание" value={adminCreate.description} onChange={(e) => setAdminCreate({ ...adminCreate, description: e.target.value })} />
            <input type="number" placeholder="Цена" value={adminCreate.price} onChange={(e) => setAdminCreate({ ...adminCreate, price: e.target.value })} />
            <input type="number" placeholder="Остаток" value={adminCreate.stock} onChange={(e) => setAdminCreate({ ...adminCreate, stock: e.target.value })} />
            <button disabled={busy} onClick={createProduct}>Создать</button>
          </article>

          <article className="card">
            <h3>Обновить товар</h3>
            <input placeholder="ID" value={adminUpdate.id} onChange={(e) => setAdminUpdate({ ...adminUpdate, id: e.target.value })} />
            <input placeholder="Название" value={adminUpdate.name} onChange={(e) => setAdminUpdate({ ...adminUpdate, name: e.target.value })} />
            <input placeholder="Категория" value={adminUpdate.category} onChange={(e) => setAdminUpdate({ ...adminUpdate, category: e.target.value })} />
            <input placeholder="Описание" value={adminUpdate.description} onChange={(e) => setAdminUpdate({ ...adminUpdate, description: e.target.value })} />
            <input type="number" placeholder="Цена" value={adminUpdate.price} onChange={(e) => setAdminUpdate({ ...adminUpdate, price: e.target.value })} />
            <input type="number" placeholder="Остаток" value={adminUpdate.stock} onChange={(e) => setAdminUpdate({ ...adminUpdate, stock: e.target.value })} />
            <button disabled={busy} onClick={updateProduct}>Обновить</button>
          </article>

          <article className="card">
            <h3>Удалить товар</h3>
            <input placeholder="ID" value={deleteProductID} onChange={(e) => setDeleteProductID(e.target.value)} />
            <button className="danger" disabled={busy} onClick={removeProduct}>Удалить</button>
          </article>
        </div>

        <div className="stack top-gap">
          <h3>Заказы всех пользователей</h3>
          {adminOrders.length === 0 ? <div className="empty">Список пуст.</div> : null}
          {adminOrders.map((order) => (
            <div className="line" key={`admin_${order.id}`}>
              <span>#{order.id} / {order.status} / {toDate(order.created_at)}</span>
              <strong>{toCurrency(order.total_amount)}</strong>
            </div>
          ))}
        </div>
      </section>
    );
  }

  function renderAuth() {
    return (
      <section className="panel">
        <h2>{isAuth ? "Профиль" : "Авторизация"}</h2>
        <div className="auth-grid">
          <article className="card">
            <h3>Регистрация + вход</h3>
            <input placeholder="Email" value={authForm.email} onChange={(e) => setAuthForm({ ...authForm, email: e.target.value })} />
            <input type="password" placeholder="Пароль" value={authForm.password} onChange={(e) => setAuthForm({ ...authForm, password: e.target.value })} />
            <input placeholder="Пол" value={authForm.gender} onChange={(e) => setAuthForm({ ...authForm, gender: e.target.value })} />
            <input type="number" placeholder="Возраст" value={authForm.age} onChange={(e) => setAuthForm({ ...authForm, age: e.target.value })} />
            <button disabled={busy} onClick={registerAndLogin}>Зарегистрироваться</button>
          </article>

          <article className="card">
            <h3>Вход</h3>
            <input placeholder="Email" value={authForm.email} onChange={(e) => setAuthForm({ ...authForm, email: e.target.value })} />
            <input type="password" placeholder="Пароль" value={authForm.password} onChange={(e) => setAuthForm({ ...authForm, password: e.target.value })} />
            <button disabled={busy} onClick={login}>Войти</button>
            <button className="ghost" disabled={busy || !refreshToken} onClick={refreshSession}>Обновить токен</button>
            {isAuth ? <button className="ghost" disabled={busy} onClick={loadProfile}>Обновить профиль</button> : null}
            {isAuth ? <button className="danger" disabled={busy} onClick={logout}>Выйти</button> : null}
          </article>

          <article className="card">
            <h3>Текущий профиль</h3>
            <p><strong>Email:</strong> {currentUser?.email || "-"}</p>
            <p><strong>Роль:</strong> {roleTitle(role)}</p>
            <p><strong>ID:</strong> {currentUser?.id || "-"}</p>
            <p><strong>Возраст:</strong> {currentUser?.age || "-"}</p>
          </article>
        </div>
      </section>
    );
  }

  function renderContent() {
    if (screen === "cart") return renderCart();
    if (screen === "orders") return renderOrders();
    if (screen === "admin") return renderAdmin();
    if (screen === "auth") return renderAuth();
    return renderCatalog();
  }

  return (
    <main className="app">
      <header className="header">
        <div>
          <p className="kicker">E-Commerce Playground</p>
          <h1>Shop Console</h1>
        </div>

        <div className="row top-links">
          {nav.map((item) => (
            <button
              key={item.key}
              className={screen === item.key ? "active" : "ghost"}
              disabled={!item.enabled || busy}
              onClick={() => setScreen(item.key)}
            >
              {item.label}
            </button>
          ))}
        </div>
      </header>

      <section className={`notice ${status.startsWith("Ошибка") ? "error" : "ok"}`}>
        <span>{status}</span>
        <span>Роль: {roleTitle(role)}</span>
      </section>

      {renderContent()}
    </main>
  );
}
