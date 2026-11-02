"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";

export default function OrdersPage() {
  const [orders, setOrders] = useState<any[]>([]);
  useEffect(() => { api("/admin/orders").then(setOrders).catch(() => {}); }, []);
  return (
    <main className="max-w-6xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">All Orders</h1>
      <div className="bg-white rounded shadow overflow-x-auto">
        <table className="w-full text-sm">
          <thead className="bg-gray-100 text-left"><tr><th className="p-3">Order</th><th className="p-3">Status</th><th className="p-3">Payment</th><th className="p-3">Total</th></tr></thead>
          <tbody>
            {orders.map((o) => (
              <tr key={o.id} className="border-t"><td className="p-3">{o.order_number}</td><td className="p-3">{o.status}</td><td className="p-3">{o.payment_status}</td><td className="p-3">₹{o.total}</td></tr>
            ))}
            {orders.length === 0 && <tr><td className="p-3 text-gray-500" colSpan={4}>No orders.</td></tr>}
          </tbody>
        </table>
      </div>
    </main>
  );
}