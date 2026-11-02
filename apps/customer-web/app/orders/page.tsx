"use client";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import Link from "next/link";

export default function OrdersPage() {
  const [orders, setOrders] = useState<any[]>([]);
  useEffect(() => { api("/orders").then(setOrders).catch(() => setOrders([])); }, []);

  return (
    <main className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4">My Orders</h1>
      {orders.length === 0 && <p>No orders yet.</p>}
      {orders.map((o) => (
        <Link key={o.id} href={`/orders/${o.id}`} className="block bg-white rounded shadow p-4 mb-2 hover:shadow-md">
          <div className="flex justify-between">
            <span className="font-medium">{o.order_number}</span>
            <span className="text-sm text-gray-500">{o.status}</span>
          </div>
          <div className="text-sm text-gray-600 mt-1">₹{o.total} · {o.payment_status}</div>
        </Link>
      ))}
    </main>
  );
}