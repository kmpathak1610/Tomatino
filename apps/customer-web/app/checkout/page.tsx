"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import { useRouter } from "next/navigation";

export default function CheckoutPage() {
  const router = useRouter();
  const [cart, setCart] = useState<any>(null);
  const [address, setAddress] = useState("");
  const [placing, setPlacing] = useState(false);

  useEffect(() => { api("/cart").then(setCart); }, []);

  const placeOrder = async () => {
    setPlacing(true);
    try {
      const items = (cart?.items || []).map((i: any) => ({ menu_item_id: i.menu_item_id, quantity: i.quantity, unit_price: i.unit_price }));
      const order = await api("/orders", {
        method: "POST",
        body: JSON.stringify({
          restaurant_id: cart.restaurant_id,
          branch_id: cart.restaurant_id,
          address_id: null,
          payment_method: "COD",
          items,
        }),
      });
      await api("/cart", { method: "DELETE" });
      router.push(`/orders/${order.order_id}`);
    } catch (e: any) {
      alert("Order failed: " + e.message);
    } finally {
      setPlacing(false);
    }
  };

  const items = cart?.items || [];
  const subtotal = items.reduce((s: number, i: any) => s + i.unit_price * i.quantity, 0);

  return (
    <main className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">Checkout</h1>
      <div className="bg-white rounded shadow p-4 mb-4">
        <h2 className="font-semibold mb-2">Order Summary</h2>
        {items.map((i: any) => (
          <div key={i.id} className="flex justify-between text-sm py-1">
            <span>{i.name} × {i.quantity}</span>
            <span>₹{(i.unit_price * i.quantity).toFixed(2)}</span>
          </div>
        ))}
        <div className="border-t mt-2 pt-2 flex justify-between font-semibold">
          <span>Total (incl. fees)</span><span>₹{(subtotal + 30 + subtotal * 0.05).toFixed(2)}</span>
        </div>
      </div>
      <div className="bg-white rounded shadow p-4 mb-4">
        <label className="block text-sm font-medium mb-1">Delivery Address</label>
        <textarea value={address} onChange={(e) => setAddress(e.target.value)} className="w-full border rounded p-2" rows={3} placeholder="Enter delivery address" />
      </div>
      <button onClick={placeOrder} disabled={placing || items.length === 0} className="w-full bg-orange-600 text-white py-2 rounded disabled:opacity-50">
        {placing ? "Placing..." : "Place Order (COD)"}
      </button>
    </main>
  );
}