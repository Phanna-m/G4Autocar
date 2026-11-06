"use client";

import {
  ArrowDownRight,
  ArrowRight,
  ArrowUpRight,
  CarFront,
  Check,
  ChevronDown,
  CircleHelp,
  Heart,
  LoaderCircle,
  Menu,
  Search,
  ShieldCheck,
  ShoppingBag,
  Sparkles,
  X,
  Zap,
} from "lucide-react";
import { useEffect, useState, type FormEvent } from "react";
import {
  apiRequest,
  Branch,
  Category,
  formatMoney,
  Product,
  productImage,
  unwrapData,
} from "@/lib/api";

type CartLine = { product: Product; quantity: number };
type OrderResponse = {
  message: string;
  data: { order_no: string; delivery_method: string };
};

const categoryIcons: Record<string, typeof CarFront> = {
  GASOLINE_VEHICLE: CarFront,
  ELECTRIC_VEHICLE: Zap,
};

function ProductArtwork({ product }: { product: Product }) {
  const image = productImage(product.images);
  if (image) {
    return (
      // Product images are supplied by the catalog, so use a regular image element.
      // eslint-disable-next-line @next/next/no-img-element
      <img className="product-photo" src={image} alt={product.name} />
    );
  }
  const Icon = product.category_type === "ELECTRIC_VEHICLE" ? Zap : CarFront;
  return (
    <div className={`product-art product-art-${(product.id.charCodeAt(0) || 0) % 4}`}>
      <span className="art-glow" />
      <Icon aria-hidden="true" size={102} strokeWidth={1.1} />
      <span className="art-label">{product.brand || "G4 AUTO CAR"}</span>
    </div>
  );
}

function formatCategory(type: string): string {
  return type
    .split("_")
    .map((part) => part.charAt(0) + part.slice(1).toLowerCase())
    .join(" ");
}

export default function HomePage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [branches, setBranches] = useState<Branch[]>([]);
  const [selectedCategory, setSelectedCategory] = useState("all");
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [cart, setCart] = useState<CartLine[]>([]);
  const [wishlistIds, setWishlistIds] = useState<string[]>([]);
  const [wishlistLoaded, setWishlistLoaded] = useState(false);
  const [wishlistOnly, setWishlistOnly] = useState(false);
  const [conditionFilter, setConditionFilter] = useState("all");
  const [sortOrder, setSortOrder] = useState("newest");
  const [cartOpen, setCartOpen] = useState(false);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [retryKey, setRetryKey] = useState(0);
  const [catalogError, setCatalogError] = useState("");
  const [notice, setNotice] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [orderConfirmation, setOrderConfirmation] = useState<{
    orderNo: string;
    deliveryMethod: string;
  } | null>(null);
  const [customer, setCustomer] = useState({
    name: "",
    phone: "",
    email: "",
    branchId: "",
    payment: "FULL_PAYMENT_CASH",
    deliveryMethod: "PICKUP",
    deliveryAddress: "",
  });

  useEffect(() => {
    try {
      const saved = window.localStorage.getItem("g4-auto-car-wishlist");
      if (saved) {
        const parsed: unknown = JSON.parse(saved);
        if (Array.isArray(parsed) && parsed.every((id) => typeof id === "string")) {
          setWishlistIds(parsed);
        }
      }
    } catch {
      window.localStorage.removeItem("g4-auto-car-wishlist");
    } finally {
      setWishlistLoaded(true);
    }
  }, []);

  useEffect(() => {
    if (!wishlistLoaded) return;
    window.localStorage.setItem("g4-auto-car-wishlist", JSON.stringify(wishlistIds));
  }, [wishlistIds, wishlistLoaded]);

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedSearch(search), 250);
    return () => window.clearTimeout(timer);
  }, [search]);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setCatalogError("");
    const query = new URLSearchParams();
    if (selectedCategory !== "all") query.set("category", selectedCategory);
    if (debouncedSearch.trim()) query.set("search", debouncedSearch.trim());
    const suffix = query.size ? `?${query.toString()}` : "";

    Promise.all([
      apiRequest<{ data: Product[] }>(`/products${suffix}`),
      apiRequest<{ data: Category[] }>("/categories"),
      apiRequest<{ data: Branch[] }>("/branches"),
    ])
      .then(([productResponse, categoryResponse, branchResponse]) => {
        if (!active) return;
        setProducts(unwrapData(productResponse) ?? []);
        setCategories(unwrapData(categoryResponse) ?? []);
        const activeBranches = (unwrapData(branchResponse) ?? []).filter(
          (branch) => branch.is_active,
        );
        setBranches(activeBranches);
        setCustomer((current) => ({
          ...current,
          branchId: current.branchId || activeBranches[0]?.id || "",
        }));
      })
      .catch((error: unknown) => {
        if (active) {
          setCatalogError(
            error instanceof Error ? error.message : "Could not load the catalog.",
          );
        }
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [selectedCategory, debouncedSearch, retryKey]);

  const cartCount = cart.reduce((total, line) => total + line.quantity, 0);
  const cartTotal = cart.reduce(
    (total, line) => total + line.product.base_price * line.quantity,
    0,
  );
  const visibleProducts = products
    .filter((product) => !wishlistOnly || wishlistIds.includes(product.id))
    .filter((product) => conditionFilter === "all" || product.condition.toLowerCase() === conditionFilter)
    .sort((a, b) => {
      if (sortOrder === "price-low") return a.base_price - b.base_price;
      if (sortOrder === "price-high") return b.base_price - a.base_price;
      if (sortOrder === "name") return a.name.localeCompare(b.name);
      return 0;
    });

  function addToCart(product: Product) {
    setCart((current) => {
      const existing = current.find((line) => line.product.id === product.id);
      if (existing) {
        return current.map((line) =>
          line.product.id === product.id
            ? { ...line, quantity: line.quantity + 1 }
            : line,
        );
      }
      return [...current, { product, quantity: 1 }];
    });
    setNotice(`${product.name} added to your enquiry.`);
    window.setTimeout(() => setNotice(""), 2600);
  }

  function updateQuantity(productId: string, delta: number) {
    setCart((current) =>
      current
        .map((line) =>
          line.product.id === productId
            ? { ...line, quantity: line.quantity + delta }
            : line,
        )
        .filter((line) => line.quantity > 0),
    );
  }

  function toggleWishlist(product: Product) {
    setWishlistIds((current) =>
      current.includes(product.id)
        ? current.filter((id) => id !== product.id)
        : [...current, product.id],
    );
  }

  function showWishlist() {
    setWishlistOnly((value) => !value);
    setSelectedCategory("all");
    setSearch("");
    setConditionFilter("all");
  }

  async function submitOrder(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!cart.length || !customer.branchId) return;
    setSubmitting(true);
    setNotice("");
    try {
      const response = await apiRequest<OrderResponse>("/orders", {
        method: "POST",
        body: JSON.stringify({
          branch_id: customer.branchId,
          customer_name: customer.name.trim(),
          customer_phone: customer.phone.trim(),
          customer_email: customer.email.trim(),
          payment_method: customer.payment,
          delivery_method: customer.deliveryMethod,
          delivery_address: customer.deliveryAddress.trim(),
          items: cart.map(({ product, quantity }) => ({
            product_id: product.id,
            quantity,
          })),
        }),
      });
      setCart([]);
      setCartOpen(false);
      setOrderConfirmation({
        orderNo: response.data.order_no,
        deliveryMethod: response.data.delivery_method,
      });
    } catch (error) {
      setNotice(error instanceof Error ? error.message : "Could not place your order.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main>
      <div className="announcement">
        <Sparkles size={14} />
        <span>Find a ride that moves you. Explore our latest collection.</span>
        <a href="#catalog">
          Shop now <ArrowRight size={14} />
        </a>
      </div>

      <header className="site-header">
        <a className="brand-lockup" href="#" aria-label="G4 Auto Car home">
          <span className="brand-mark"><CarFront size={22} /></span>
          <span>G4 <span>Auto Car</span><small>DRIVE YOUR NEXT</small></span>
        </a>
        <nav className={mobileNavOpen ? "main-nav mobile-open" : "main-nav"} aria-label="Main navigation">
          <a className="active" href="#catalog" onClick={() => setMobileNavOpen(false)}>Vehicles</a>
          <a href="#why-us" onClick={() => setMobileNavOpen(false)}>Why G4 Auto Car</a>
          <a href="#branches" onClick={() => setMobileNavOpen(false)}>Our branches</a>
        </nav>
        <div className="header-actions">
          <a className="admin-link" href="/admin">Dealer portal <ArrowUpRight size={15} /></a>
          <button className="icon-button cart-button" onClick={() => setCartOpen(true)} aria-label={`Open enquiry basket, ${cartCount} items`}>
            <ShoppingBag size={19} />
            <span>Enquiry</span>
            <b>{cartCount}</b>
          </button>
          <button className="icon-button wishlist-button" onClick={() => { showWishlist(); document.getElementById("catalog")?.scrollIntoView({ behavior: "smooth" }); }} aria-pressed={wishlistOnly} aria-label={`Show wishlist, ${wishlistIds.length} items`}>
            <Heart size={17} fill={wishlistOnly ? "currentColor" : "none"} />
            <span>Wishlist</span>
            <b>{wishlistIds.length}</b>
          </button>
          <button className="mobile-menu" onClick={() => setMobileNavOpen((open) => !open)} aria-label={mobileNavOpen ? "Close menu" : "Open menu"} aria-expanded={mobileNavOpen}>
            {mobileNavOpen ? <X size={21} /> : <Menu size={21} />}
          </button>
        </div>
      </header>

      <section className="hero">
        <div className="hero-copy">
          <div className="eyebrow"><span /> YOUR JOURNEY STARTS HERE</div>
          <h1>Find the one<br />that feels like <em>you.</em></h1>
          <p>From the open road to the everyday commute, discover vehicles chosen for the way you live.</p>
          <div className="hero-actions">
            <a className="button button-dark" href="#catalog">Explore collection <ArrowRight size={17} /></a>
            <a className="text-link" href="#why-us">Why shop with us <ArrowDownRight size={16} /></a>
          </div>
          <div className="hero-proof">
            <div className="avatar-stack"><span>V</span><span>H</span><span>+</span></div>
            <div><strong>Made for real journeys</strong><small>Helpful people. Honest choices.</small></div>
          </div>
        </div>
        <div className="hero-visual">
          <div className="hero-photo-stage">
            <span className="hero-stamp"><Sparkles size={14} /> THE NEXT CHAPTER</span>
            {/* Showcase-only image, intentionally not linked to any catalog vehicle. */}
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="hero-car-photo" src="/images/showcase-car.png" alt="Sports car on a coastal road at sunset" />
            <div className="hero-image-shade" />
            <div className="hero-floating-card">
              <span className="floating-icon"><ShieldCheck size={18} /></span>
              <span><strong>A ride you can trust</strong><small>Quality checked, always</small></span>
              <Check size={17} className="floating-check" />
            </div>
          </div>
          <span className="visual-caption">GOOD THINGS ARE IN MOTION <ArrowUpRight size={13} /></span>
        </div>
        <div className="hero-bottom">
          <span>01 — MADE FOR YOUR NEXT CHAPTER</span>
          <span className="hero-bottom-line" />
          <span>{products.length ? `${products.length} VEHICLES TO DISCOVER` : "SCROLL TO EXPLORE"} <ChevronDown size={14} /></span>
        </div>
      </section>

      <section className="trust-strip" id="why-us">
        <div><span className="trust-icon"><ShieldCheck size={19} /></span><span><strong>Quality you can feel</strong><small>Every vehicle, thoughtfully selected</small></span></div>
        <div><span className="trust-icon"><CircleHelp size={19} /></span><span><strong>People who get it</strong><small>Friendly help, no pressure</small></span></div>
        <div><span className="trust-icon"><Zap size={19} /></span><span><strong>Easy from day one</strong><small>Simple, clear buying experience</small></span></div>
      </section>

      <section className="catalog-section" id="catalog">
        <div className="section-heading">
          <div>
            <div className="eyebrow"><span /> THE COLLECTION</div>
            <h2>Good rides, <em>great stories.</em></h2>
            <p>A little something for every kind of road ahead.</p>
          </div>
          <label className="search-field">
            <Search size={18} />
            <input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search makes, models..." aria-label="Search vehicles" />
            {search && <button onClick={() => setSearch("")} aria-label="Clear search"><X size={15} /></button>}
          </label>
        </div>

        <div className="category-row" aria-label="Filter by category">
          <button className={selectedCategory === "all" ? "category-chip selected" : "category-chip"} onClick={() => setSelectedCategory("all")}>Everything <span>{products.length}</span></button>
          {categories.map((category) => {
            const Icon = categoryIcons[category.type] ?? CarFront;
            return (
              <button key={category.id} className={selectedCategory === category.slug ? "category-chip selected" : "category-chip"} onClick={() => setSelectedCategory(category.slug)}>
                <Icon size={16} /> {category.name}
              </button>
            );
          })}
        </div>
        <div className="catalog-controls">
          <button className={wishlistOnly ? "category-chip selected" : "category-chip"} onClick={showWishlist}>
            <Heart size={14} /> Wishlist ({wishlistIds.length})
          </button>
          <label>Condition
            <select value={conditionFilter} onChange={(event) => setConditionFilter(event.target.value)}>
              <option value="all">All conditions</option>
              <option value="new">New</option>
              <option value="used">Used</option>
              <option value="certified">Certified</option>
            </select>
          </label>
          <label>Sort by
            <select value={sortOrder} onChange={(event) => setSortOrder(event.target.value)}>
              <option value="newest">Newest</option>
              <option value="price-low">Price: low to high</option>
              <option value="price-high">Price: high to low</option>
              <option value="name">Name</option>
            </select>
          </label>
        </div>

        {catalogError ? (
          <div className="empty-state error-state">
            <span className="empty-icon"><CircleHelp size={25} /></span>
            <h3>We couldn&apos;t reach the collection.</h3>
            <p>{catalogError}. Check that the backend is running, then try again.</p>
            <button className="button button-dark" onClick={() => setRetryKey((key) => key + 1)}>Try again <ArrowRight size={16} /></button>
          </div>
        ) : loading ? (
          <div className="loading-state"><LoaderCircle className="spin" size={22} /> Finding your next ride...</div>
        ) : visibleProducts.length ? (
          <>
            <div className="catalog-grid">
              {visibleProducts.map((product, index) => (
                <article className="product-card" key={product.id}>
                  <div className="product-image-wrap">
                    <ProductArtwork product={product} />
                    <span className="product-tag">{product.condition || "NEW"}</span>
                    <button className={`wishlist-toggle${wishlistIds.includes(product.id) ? " active" : ""}`} onClick={() => toggleWishlist(product)} aria-label={`${wishlistIds.includes(product.id) ? "Remove" : "Add"} ${product.name} ${wishlistIds.includes(product.id) ? "from" : "to"} wishlist`} aria-pressed={wishlistIds.includes(product.id)}><Heart size={16} fill={wishlistIds.includes(product.id) ? "currentColor" : "none"} /></button>
                    <button className="quick-add" onClick={() => addToCart(product)} aria-label={`Add ${product.name} to enquiry`}><ShoppingBag size={17} /></button>
                  </div>
                  <div className="product-info">
                    <div className="product-meta"><span>{product.brand || product.category_name || "G4 AUTO CAR"}</span><span>{String(index + 1).padStart(2, "0")}</span></div>
                    <h3>{product.name}</h3>
                    <div className="product-subtitle">{product.model}{product.year ? ` · ${product.year}` : ""}</div>
                    <div className="product-bottom"><strong>{formatMoney(product.base_price)}</strong><button onClick={() => addToCart(product)}>Add to enquiry <ArrowRight size={15} /></button></div>
                  </div>
                </article>
              ))}
            </div>
            <div className="catalog-footnote"><span><span className="tiny-dot" /> {visibleProducts.length} vehicles to explore</span><span>Prices shown in USD</span></div>
          </>
        ) : (
          <div className="empty-state">
            <span className="empty-icon"><Search size={24} /></span>
            <h3>{wishlistOnly && !wishlistIds.length ? "Your wishlist is empty." : "No rides found just yet."}</h3>
            <p>{wishlistOnly && !wishlistIds.length ? "Save a vehicle with the heart button and it will stay here on this device." : "Try a different search or browse another category."}</p>
            <button className="button button-outline" onClick={() => { setSearch(""); setSelectedCategory("all"); setWishlistOnly(false); setConditionFilter("all"); }}>Clear filters</button>
          </div>
        )}
      </section>

      <section className="feature-banner" id="branches">
        <div className="feature-copy">
          <div className="eyebrow eyebrow-light"><span /> A BETTER WAY TO BUY</div>
          <h2>Good people.<br /><em>Good places.</em></h2>
          <p>Visit one of our branches and meet a team that&apos;s here for your journey—not just the sale.</p>
          <a className="button button-light" href="#branch-list">Find your branch <ArrowRight size={16} /></a>
        </div>
        <div className="feature-art" aria-hidden="true">
          <div className="feature-circle feature-circle-back" />
          <div className="feature-circle feature-circle-front" />
          <div className="feature-star">✳</div>
          <span className="feature-word">HERE<br />FOR THE<br /><em>RIDE.</em></span>
          <div className="feature-doodle">↗</div>
        </div>
        <div className="branch-list" id="branch-list">
          {branches.slice(0, 3).map((branch, index) => (
            <div className="branch-line" key={branch.id}><span>0{index + 1}</span><strong>{branch.name}</strong><small>{branch.city}</small><ArrowUpRight size={16} /></div>
          ))}
          {!branches.length && <div className="branch-line"><span>01</span><strong>Local branches</strong><small>Ready to welcome you</small><ArrowUpRight size={16} /></div>}
        </div>
      </section>

      <section className="closing-cta">
        <div className="eyebrow"><span /> YOUR NEXT CHAPTER</div>
        <h2>Ready when <em>you are.</em></h2>
        <p>Start with a browse. Stay for the feeling.</p>
        <a className="button button-dark" href="#catalog">Find your next ride <ArrowRight size={16} /></a>
      </section>

      <footer className="site-footer">
        <a className="brand-lockup footer-brand" href="#"><span className="brand-mark"><CarFront size={20} /></span><span>G4 <span>Auto Car</span><small>DRIVE YOUR NEXT</small></span></a>
        <span>Thoughtful rides. Real people. Your next chapter.</span>
        <span>© {new Date().getFullYear()} G4 Auto Car</span>
      </footer>

      {notice && <div className="toast-message" role="status"><Check size={17} />{notice}<button onClick={() => setNotice("")} aria-label="Dismiss notification"><X size={15} /></button></div>}

      {cartOpen && (
        <div className="modal-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setCartOpen(false); }}>
          <section className="cart-panel" role="dialog" aria-modal="true" aria-labelledby="cart-title">
            <div className="cart-header"><div><div className="eyebrow"><span /> YOUR SHORTLIST</div><h2 id="cart-title">Enquiry basket <small>({cartCount})</small></h2></div><button className="icon-button close-button" onClick={() => setCartOpen(false)} aria-label="Close basket"><X size={19} /></button></div>
            {!cart.length ? (
              <div className="cart-empty"><ShoppingBag size={30} /><h3>Your shortlist is waiting.</h3><p>Add a vehicle to start an enquiry.</p><button className="button button-dark" onClick={() => setCartOpen(false)}>Explore vehicles <ArrowRight size={16} /></button></div>
            ) : (
              <>
                <div className="cart-items">
                  {cart.map(({ product, quantity }) => (
                    <div className="cart-item" key={product.id}>
                      <div className="cart-item-art"><CarFront size={25} /></div>
                      <div className="cart-item-main"><strong>{product.name}</strong><small>{product.brand} · {formatMoney(product.base_price)}</small><div className="quantity-control"><button onClick={() => updateQuantity(product.id, -1)} aria-label="Decrease quantity">−</button><span>{quantity}</span><button onClick={() => updateQuantity(product.id, 1)} aria-label="Increase quantity">+</button></div></div>
                      <button className="remove-item" onClick={() => updateQuantity(product.id, -quantity)} aria-label={`Remove ${product.name}`}><X size={16} /></button>
                    </div>
                  ))}
                </div>
                <form className="checkout-form" onSubmit={submitOrder}>
                  <div className="cart-total"><span>Estimated total</span><strong>{formatMoney(cartTotal)}</strong></div>
                  <div className="form-grid">
                    <label>Your name<input required value={customer.name} onChange={(event) => setCustomer({ ...customer, name: event.target.value })} placeholder="Full name" autoComplete="name" /></label>
                    <label>Phone number<input required value={customer.phone} onChange={(event) => setCustomer({ ...customer, phone: event.target.value })} placeholder="+855 ..." autoComplete="tel" /></label>
                    <label>Email <span>(optional)</span><input type="email" value={customer.email} onChange={(event) => setCustomer({ ...customer, email: event.target.value })} placeholder="you@example.com" autoComplete="email" /></label>
                    <label>Preferred branch<select required value={customer.branchId} onChange={(event) => setCustomer({ ...customer, branchId: event.target.value })}><option value="">Choose a branch</option>{branches.map((branch) => <option value={branch.id} key={branch.id}>{branch.name} · {branch.city}</option>)}</select></label>
                    <label className="full-field">Payment preference<select value={customer.payment} onChange={(event) => setCustomer({ ...customer, payment: event.target.value })}><option value="FULL_PAYMENT_CASH">Full payment · Cash</option><option value="FULL_PAYMENT_KHQR">Full payment · KHQR</option><option value="INSTALLMENT">Installment plan</option><option value="TRADE_IN">Trade-in</option></select></label>
                    <label className="full-field">Fulfilment<select value={customer.deliveryMethod} onChange={(event) => setCustomer({ ...customer, deliveryMethod: event.target.value })}><option value="PICKUP">Collect at selected branch</option><option value="DELIVERY">Delivery by G4 Auto Car</option></select></label>
                    {customer.deliveryMethod === "DELIVERY" && <label className="full-field">Delivery address<textarea required value={customer.deliveryAddress} onChange={(event) => setCustomer({ ...customer, deliveryAddress: event.target.value })} placeholder="House number, street, district, city / province" rows={3} /></label>}
                  </div>
                  <p className="checkout-note">{customer.deliveryMethod === "DELIVERY" ? "Delivery fee is not included yet. Our team will confirm the location-based fee and final amount before payment." : "Our team will confirm availability and next steps with you."}</p>
                  <button className="button button-dark checkout-button" type="submit" disabled={submitting || !branches.length}>{submitting ? <><LoaderCircle className="spin" size={17} /> Sending...</> : <>Send enquiry <ArrowRight size={16} /></>}</button>
                  {!branches.length && <p className="form-error">Branch information is unavailable. Please try again later.</p>}
                </form>
              </>
            )}
          </section>
        </div>
      )}
      {orderConfirmation && (
        <div className="modal-backdrop confirmation-backdrop" onMouseDown={(event) => { if (event.target === event.currentTarget) setOrderConfirmation(null); }}>
          <section className="confirmation-card" role="dialog" aria-modal="true" aria-labelledby="confirmation-title">
            <span className="confirmation-icon"><Check size={25} /></span>
            <div className="eyebrow"><span /> ORDER RECEIVED</div>
            <h2 id="confirmation-title">Thank you for choosing G4 Auto Car.</h2>
            <p>Your request is saved. Our team will contact you to confirm vehicle availability{orderConfirmation.deliveryMethod === "DELIVERY" ? " and your delivery fee" : ""} before payment.</p>
            <div className="confirmation-number"><span>Order number</span><strong>{orderConfirmation.orderNo}</strong></div>
            <button className="button button-dark checkout-button" onClick={() => setOrderConfirmation(null)}>Continue shopping <ArrowRight size={16} /></button>
          </section>
        </div>
      )}
    </main>
  );
}
