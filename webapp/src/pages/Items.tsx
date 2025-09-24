import { useEffect, useState } from "react";
import { json, apiFetch } from "../api";
import { useAuth } from "../auth";

type Item = { id: number; name: string; price: number; created_at: string };

export default function Items() {
    const { user, logout } = useAuth();
    const [items, setItems] = useState<Item[]>([]);
    const [name, setName] = useState("");
    const [price, setPrice] = useState<number>(0);
    const [editId, setEditId] = useState<number | null>(null);
    const [editName, setEditName] = useState("");
    const [editPrice, setEditPrice] = useState<number>(0);

    function toast(m: string) { alert(m); }

    async function load() {
        try { setItems(await json<Item[]>("/items/")); }
        catch (e: any) { toast(e?.error ?? "Load failed"); }
    }
    useEffect(() => { void load(); }, []);

    async function add(e: React.FormEvent) {
        e.preventDefault();
        try {
            await json("/items/", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name, price: Number(price) })
            });
            setName(""); setPrice(0); await load();
        } catch (e: any) { toast(e?.error ?? "Create failed"); }
    }

    function startEdit(i: Item) { setEditId(i.id); setEditName(i.name); setEditPrice(i.price); }
    function cancelEdit() { setEditId(null); setEditName(""); setEditPrice(0); }

    async function saveEdit() {
        if (editId == null) return;
        try {
            await json(`/items/${editId}`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name: editName, price: Number(editPrice) })
            });
            cancelEdit(); await load();
        } catch (e: any) { toast(e?.error ?? "Update failed"); }
    }

    async function remove(id: number) {
        const r = await apiFetch(`/items/${id}`, { method: "DELETE" });
        if (r.status === 403) { toast("Forbidden: admin required"); return; }
        if (r.status !== 204) {
            const j = await r.json().catch(() => null);
            toast(j?.error ?? `Delete failed ${r.status}`); return;
        }
        await load();
    }

    return (
        <div style={{ maxWidth: 920, margin: "40px auto", fontFamily: "system-ui" }}>
            <header style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                <h1>Items</h1>
                <div>
                    <span style={{ marginRight: 8 }}>role: <b>{user?.role}</b></span>
                    <button onClick={logout}>Logout</button>
                </div>
            </header>

            <form onSubmit={add} style={{ display: "flex", gap: 8, marginBottom: 16 }}>
                <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Name" required />
                <input type="number" step="0.01" value={price} onChange={(e) => setPrice(Number(e.target.value))} placeholder="Price" />
                <button type="submit">Add</button>
            </form>

            <ul style={{ listStyle: "none", padding: 0 }}>
                {items.map((i) => (
                    <li key={i.id} style={{ display: "flex", gap: 8, alignItems: "center", borderBottom: "1px solid #eee", padding: "6px 0" }}>
                        {editId === i.id ? (
                            <>
                                <input value={editName} onChange={(e) => setEditName(e.target.value)} />
                                <input type="number" step="0.01" value={editPrice} onChange={(e) => setEditPrice(Number(e.target.value))} />
                                <button onClick={saveEdit} type="button">Save</button>
                                <button onClick={cancelEdit} type="button">Cancel</button>
                            </>
                        ) : (
                            <>
                                <span>#{i.id}</span>
                                <span style={{ flex: 1 }}>{i.name}</span>
                                <span>{i.price.toFixed(2)}</span>
                                <span>{new Date(i.created_at).toLocaleString()}</span>
                                <button onClick={() => startEdit(i)} type="button">Edit</button>
                                {user?.role === "admin" && (
                                    <button onClick={() => remove(i.id)} type="button">Delete</button>
                                )}
                            </>
                        )}
                    </li>
                ))}
            </ul>
        </div>
    );
}
