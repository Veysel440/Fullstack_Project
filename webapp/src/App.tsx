import { useEffect, useState } from "react";

type Item = { id: number; name: string; price: number; created_at: string };
const API = (import.meta.env.VITE_API_URL as string) ?? "http://localhost:8080";

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));
async function fetchRetry(input: RequestInfo, init?: RequestInit) {
    for (let a = 0; a < 4; a++) {
        const res = await fetch(input, init);
        if (res.status !== 429) return res;
        await sleep(200 * Math.pow(2, a));
    }
    return fetch(input, init);
}
function toast(msg: string) {
    alert(msg);
}

export default function App() {
    const [items, setItems] = useState<Item[]>([]);
    const [name, setName] = useState("");
    const [price, setPrice] = useState<number>(0);
    const [editId, setEditId] = useState<number | null>(null);
    const [editName, setEditName] = useState("");
    const [editPrice, setEditPrice] = useState<number>(0);

    async function load() {
        const res = await fetchRetry(`${API}/items/`);
        if (!res.ok) {
            const j = await res.json().catch(() => null);
            toast(j?.error ?? `Load failed ${res.status}`);
            return;
        }
        setItems(await res.json());
    }

    async function add(e: React.FormEvent) {
        e.preventDefault();
        const res = await fetchRetry(`${API}/items/`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name, price: Number(price) })
        });
        if (!res.ok) {
            const j = await res.json().catch(() => null);
            toast(j?.error ?? `Create failed ${res.status}`);
            return;
        }
        setName("");
        setPrice(0);
        await load();
    }

    function startEdit(i: Item) {
        setEditId(i.id);
        setEditName(i.name);
        setEditPrice(i.price);
    }
    function cancelEdit() {
        setEditId(null);
        setEditName("");
        setEditPrice(0);
    }
    async function saveEdit() {
        if (editId == null) return;
        const res = await fetchRetry(`${API}/items/${editId}`, {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name: editName, price: Number(editPrice) })
        });
        if (!res.ok) {
            const j = await res.json().catch(() => null);
            toast(j?.error ?? `Update failed ${res.status}`);
            return;
        }
        cancelEdit();
        await load();
    }
    async function remove(id: number) {
        const res = await fetchRetry(`${API}/items/${id}`, { method: "DELETE" });
        if (!res.ok && res.status !== 204) {
            const j = await res.json().catch(() => null);
            toast(j?.error ?? `Delete failed ${res.status}`);
            return;
        }
        await load();
    }

    useEffect(() => { void load(); }, []);

    return (
        <div style={{ maxWidth: 860, margin: "40px auto", fontFamily: "system-ui" }}>
            <h1>PostgreSQL Items</h1>

            <form onSubmit={add} style={{ display: "flex", gap: 8, marginBottom: 16 }}>
                <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Ad" required />
                <input
                    type="number"
                    step="0.01"
                    value={price}
                    onChange={(e) => setPrice(Number(e.target.value))}
                    placeholder="Fiyat"
                />
                <button type="submit">Ekle</button>
            </form>

            <ul style={{ listStyle: "none", padding: 0 }}>
                {items.map((i) => (
                    <li key={i.id} style={{ display: "flex", alignItems: "center", gap: 8, padding: "6px 0", borderBottom: "1px solid #eee" }}>
                        {editId === i.id ? (
                            <>
                                <input value={editName} onChange={(e) => setEditName(e.target.value)} />
                                <input
                                    type="number"
                                    step="0.01"
                                    value={editPrice}
                                    onChange={(e) => setEditPrice(Number(e.target.value))}
                                />
                                <button onClick={saveEdit} type="button">Kaydet</button>
                                <button onClick={cancelEdit} type="button">Vazgeç</button>
                            </>
                        ) : (
                            <>
                                <span>#{i.id}</span>
                                <span style={{ flex: 1 }}>{i.name}</span>
                                <span>{i.price.toFixed(2)}</span>
                                <span>{new Date(i.created_at).toLocaleString()}</span>
                                <button onClick={() => startEdit(i)} type="button">Düzenle</button>
                                <button onClick={() => remove(i.id)} type="button">Sil</button>
                            </>
                        )}
                    </li>
                ))}
            </ul>
        </div>
    );
}
