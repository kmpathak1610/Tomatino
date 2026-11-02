import "./globals.css";
import Link from "next/link";

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="min-h-screen bg-gray-50 text-gray-900">
        <nav className="bg-white shadow-sm sticky top-0 z-10">
          <div className="max-w-6xl mx-auto px-4 py-3 flex items-center justify-between">
            <Link href="/" className="text-xl font-bold text-orange-600">FoodExpress</Link>
            <div className="flex gap-4 text-sm">
              <Link href="/cart">Cart</Link>
              <Link href="/orders">Orders</Link>
              <Link href="/login">Login</Link>
            </div>
          </div>
        </nav>
        {children}
      </body>
    </html>
  );
}