"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import Link from "next/link";

export default function CartPage() {
  const [cart, setCart] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  const load = () => {
    api("/cart").then(setCart).catch(() => setCart(null)).finally(() => setLoading(false));
  };
  useEffect(load, []);

  const remove = async (id: string) => { await api(`/cart/items/${id}`, { method: "DELETE" }); load(); };

  if (loading) return <main className="p-6">Loading...</main>;

  const items = cart?.items || [];
  const subtotal = items.reduce((s: number, i: any) => s + i.unit_price * i.quantity, 0);

  return (
    <main className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Your Cart</h1>
      {items.length === 0 && <p>Cart is empty. <Link href="/" className="text-orange-600">Browse restaurants</Link></p>}
      {items.map((i: any) => (
        <div key={i.id} className="bg-white rounded shadow p-4 mb-2 flex justify-between items-center">
          <div>
            <div className="font-medium">{i.name}</div>
            <div className="text-sm text-gray-500">₹{i.unit_price} × {i.quantity}</div>
          </div>
          <button onClick={() => remove(i.id)} className="text-red-600 text-sm">Remove</button>
        </div>
      ))}
      {items.length > 0 && (
        <div className="mt-4 bg-white rounded shadow p-4">
          <div className="flex justify-between font-semibold"><span>Subtotal</span><span>₹{subtotal.toFixed(2)}</span></div>
          <Link href="/checkout" className="block mt-3 bg-orange-600 text-white text-center py-2 rounded">Proceed to Checkout</Link>
        </div>
      )}
    </main>
  );
}