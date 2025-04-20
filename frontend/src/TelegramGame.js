// lvlchess/frontend/src/TelegramGame.js
import React, { useEffect, useState } from 'react';
import ChessGame from './ChessGame';




export default function TelegramGame() {
    const [user, setUser] = useState(null);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        // Telegram.WebApp может быть ещё не инициализирован – ждём события ready:
        if (window.Telegram && window.Telegram.WebApp) {
            window.Telegram.WebApp.onEvent('mainButtonClicked', () => {});
            window.Telegram.WebApp.ready();

            const initData = window.Telegram.WebApp.initData;
            if (!initData) {
                setError('initData отсутствует');
                return;
            }

            // Отправляем на бэкенд для проверки подписи
            fetch('/api/checkInitData', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: new URLSearchParams({ initData }),
            })
                .then(res => res.json())
                .then((data) => {
                    if (!data.ok) {
                        throw new Error('Неправильная подпись initData');
                    }
                    setUser({ id: data.user_id, username: data.username });
                })
                .catch((err) => {
                    console.error(err);
                    setError(err.message);
                });
        } else {
            setError('Telegram.WebApp недоступен');
        }
    }, []);

    if (error) {
        return <div style={{ padding: 20, color: 'red' }}>Ошибка: {error}</div>;
    }
    if (!user) {
        return <div style={{ padding: 20 }}>Проверка Telegram initData…</div>;
    }

    // Коллбэк: каждый ход шлём bot’у через Telegram.WebApp.sendData
    const handleMove = move => {
        window.Telegram.WebApp.sendData(JSON.stringify({
            type: 'move',
            from: move.from,
            to:   move.to,
            promotion: move.promotion
        }));
    };
    return (
        <div style={{ padding: 20 }}>
            <h2>Добро пожаловать, {user.username || user.id}</h2>
            <ChessGame onMove={handleMove} />
        </div>
    );
}
