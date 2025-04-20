import React, { useState, useEffect, useCallback } from 'react';
import { Chess } from 'chess.js';
import { Chessboard } from 'react-chessboard';

export default function ChessGame({ onMove }) {
    // 1) Инициализируем движок и состояние FEN
    const [game] = useState(() => new Chess());
    const [fen, setFen] = useState(game.fen());
    const [status, setStatus] = useState('');

    // 2) Функция, которая вызывается при попытке перетянуть фигуру
    const onDrop = useCallback((sourceSquare, targetSquare) => {
        const move = game.move({
            from: sourceSquare,
            to:   targetSquare,
            promotion: 'q'  // всегда превращаем в ферзя
        });
        if (move === null) {
            // ход нелегальный — откат доски
            return false;
        }
        // 3) Успешный ход: обновляем FEN и статус
        setFen(game.fen());
        const newStatus = game.in_checkmate()
            ? `Мат! Победили ${move.color === 'w' ? 'Белые' : 'Чёрные'}`
            : game.in_draw()
                ? 'Ничья'
                : `Ходит ${game.turn() === 'w' ? 'Белые' : 'Чёрные'}`;
        setStatus(newStatus);

        // 4) И если нужно, сообщаем родителю (TelegramGame) о ходе
        onMove && onMove(move);

        return true;
    }, [game, onMove]);

    return (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center' }}>
            <Chessboard
                position={fen}
                onPieceDrop={onDrop}
                boardWidth={400}
            />
            <div style={{ marginTop: 16, fontSize: '1.1rem' }}>{status}</div>
        </div>
    );
}
