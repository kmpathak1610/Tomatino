"use client";
import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { api } from "@/lib/api";

export default function RestaurantDetail() {
  const { id } = useParams();
  const [restaurant, setRestaurant] = useState<any>(null);
  const [menu, setMenu] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    Promise.all([
      api(`/restaurants/${id}`),
      api(`/menu?restaurant_id=${id}`),
    ])
      .then(([r, m]) => { setRestaurant(r); setMenu(m); })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [id]);

  const addToCart = async (item: any) => {
    const token = localStorage.getItem("token");
    if (!token) { window.location.href = "/login"; return; }
    await api("/cart/items", {
      method: "POST",
      body: JSON.stringify({ restaurant_id: id, branch_id: id, menu_item_id: item.id, quantity: 1 }),
    });
    alert("Added to cart");
  };

  if (loading) return <main className="p-6">Loading...</main>;
  if (!restaurant) return <main className="p-6">Not found</main>;

  return (
    <main className="max-w-6xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-1">{restaurant.name}</h1>
      <p className="text-gray-600 mb-4">{restaurant.description}</p>
      <div className="text-sm text-gray-500 mb-6">⭐ {restaurant.avg_rating} · {(restaurant.cuisine_types || []).join(", ")}</div>
      <h2 className="text-xl font-semibold mb-3">Menu</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {menu.map((m) => (
          <div key={m.id} className="bg-white rounded shadow p-4 flex justify-between items-start">
            <div>
              <div className="font-medium">{m.is_veg ? "🟢" : "🔴"} {m.name}</div>
              <div className="text-sm text-gray-500">{m.description}</div>
              <div className="mt-1 font-semibold">₹{m.base_price}</div>
            </div>
            <button onClick={() => addToCart(m)} className="bg-orange-600 text-white px-3 py-1 rounded text-sm">Add</button>
          </div>
        ))}
      </div>
    </main>
  );
}