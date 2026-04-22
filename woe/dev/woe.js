const palette = [
    "#ca2652",
    "#79751b",
    "#a23737",
    "#2144f3",
    "#683572",
    "#0a7d4a",
    "#0e4467",
];

let sectors = [
    {color:palette[0], label:"one"},
    {color:palette[1], label:"two"},
    {color:palette[2], label:"three"},
    {color:palette[3], label:"four"},
    {color:palette[4], label:"five"},
    {color:palette[5], label:"six"},
    {color:palette[6], label:"seven"},
];
let entries = ["one", "two", "three", "four", "five", "six", "seven"];
let velocity = 0.30;
let winnerNow = null;

// Form entries
document.querySelector("#start-game").addEventListener("click", () => {
    const entries = document.querySelector("#entries-input")
        .value.split("\n")
        .map(e => e.trim())
        .filter(e => e.length > 0);

    fetch("/spin", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-Spin-ID": crypto.randomUUID(),
        },
        body: JSON.stringify({
            id: crypto.randomUUID(),
            version: "0.1.0",
            event_type: "spin.browser.wheel",
            data: { entries }
        })
    });
});

// WebSocket connection (PROD)
// const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
// const ws = new WebSocket(`${proto}//${window.location.host}/ws`);

// WebSocket connection (DEV)
const ws = new WebSocket('ws://localhost:1234/ws');

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received data: ', JSON.stringify(data));

    if (!data.entries.length || data.entries.length === 0) return;

    const timestamp = data.timestamp;
    const spinID = data.id;
    entries = data.entries;
    velocity = data.velocity;

    sectors = data.entries.map((label, i) => ({
        label,
        color: palette[i % palette.length]
    }));
    tot = sectors.length;
    arc = TAU / sectors.length;
    sectors.forEach(drawSector);

    if (data.type === "spin") { // Spin the wheel by making velocity non-zero
        winnerNow = data.spun; // Must be set before setting angVel
        if (!angVel) angVel = velocity;
        rotate();
    } else if (data.type === "sync") { // rotate to the sector at the current index
        rotate();
    }
}

const EL_spin = document.querySelector("#spin");
const ctx = document.querySelector("#wheel").getContext('2d');
const dia = ctx.canvas.width;
const rad = dia / 2;
const PI = Math.PI;
const TAU = 2 * PI;
let arc = TAU / sectors.length;
let tot = sectors.length;

const friction = 0.991; // 0.995=soft, 0.99=mid, 0.98=hard
// let angVel = 0; // Angular velocity
let angVel = velocity; // Angular velocity
let ang = 0; // Angle in radians

const getIndex = () => Math.floor(tot - ang / TAU * tot) % tot;

function drawSector(sector, i) {
    const ang = arc * i;
    ctx.save();
    // COLOR
    ctx.beginPath();
    // const gradient = ctx.createLinearGradient(0, 0, dia, dia);
    // gradient.addColorStop(0, lightenColor(sector.color));
    // gradient.addColorStop(1, sector.color);
    ctx.fillStyle = createLogGradient(ctx, sector.color);
    ctx.moveTo(rad, rad);
    ctx.arc(rad, rad, rad, ang, ang + arc);
    ctx.lineTo(rad, rad);
    ctx.fill();
    ctx.strokeStyle = 'rgba(255,255,255,0.4)';
    ctx.lineWidth = 2;
    ctx.stroke();
    // TEXT
    const words = sector.label.split(" "); // Account for multiple words
    const lineHeight = 34;
    const startY = 10 - ((words.length - 1) * lineHeight) /2;
    ctx.translate(rad, rad);
    ctx.rotate(ang + arc / 2);
    ctx.textAlign = "right";
    ctx.fillStyle = "#21d7f3";
    ctx.font = "bold 30px sans-serif";
    ctx.strokeStyle = "rgba(0,0,0,0.6)"
    ctx.lineWidth = 3;
    ctx.lineJoin = "round";
    words.forEach((word, i) => {
        ctx.strokeText(word, rad - 10, startY + i * lineHeight);
        ctx.fillText(word, rad - 10, startY + i * lineHeight);
    });
    // ctx.fillText(sector.label, rad - 10, 10);
    ctx.restore();
}

function lightenColor(hex, amount = 80) {
    const num = parseInt(hex.slice(1), 16);
    const r = Math.min(255, (num >> 16) + amount);
    const g = Math.min(255, ((num >> 8) & 0xff) + amount);
    const b = Math.min(255, (num & 0xff) + amount);
    return `rgb(${r},${g},${b})`;
}

function createLogGradient(ctx, color) {
    const gradient = ctx.createLinearGradient(0, 0, dia, dia);
    const steps = 10;
    for (let i = 0; i <= steps; i++) {
        const t = Math.log1p(i) / Math.log1p(steps); // 0..1 on a log curve
        const amount = Math.floor((1 - t) * 80);
        // const amount = Math.floor(t * 80);
        gradient.addColorStop(t, lightenColor(color, amount));
    }
    return gradient;
}

function snapWinnerNow(name) {
    const idx = sectors.findIndex(s => s.label === name);
    console.log(`snapWinnerNow: name="${name}" idx=${idx} ang=${ang.toFixed(4)} arc=${arc.toFixed(4)} tot=${tot}`);
    if (idx === -1) return;
    const targetAng = -Math.PI / 2 - (idx * arc) - (arc / 2);
    ang = ((targetAng % TAU) + TAU) % TAU;
    console.log(`snapWinnerNow: targetAng=${targetAng.toFixed(4)} finalAng=${ang.toFixed(4)}`);
    rotate();
}

function rotate() {
    // if (!sectors.length) return; // Guard when sectors haven't been received yet?
    const sector = sectors[getIndex()];
    ctx.canvas.style.transform = `rotate(${ang - PI / 2}rad)`;
    // EL_spin.textContent = !angVel ? "spin" : sector.label;
    // EL_spin.style.background = sector.color;
    EL_spin.textContent = sector.label;
    EL_spin.style.background = `radial-gradient(circle at 35% 35%, white, ${sector.color})`;
}

function frame() {
    if (!angVel) return;
    angVel *= friction; // Decrement velocity by friction

    // When stopping, snap to the winning sector
    if (angVel < 0.002) {
        angVel = 0; // Bring to stop
        snapWinnerNow(winnerNow);
        return;
    }
    ang += angVel; // Update angle
    ang %= TAU; // Normalize angle
    rotate();
}

function engine() {
    frame();
    requestAnimationFrame(engine)
}

// INIT
sectors.forEach(drawSector);
rotate(); // Initial rotation
engine(); // Start engine
EL_spin.addEventListener("click", () => {
    if (!angVel) {
        angVel = velocity;
        ws.send(JSON.stringify({
            type: "spin",
            velocity: angVel,
        }));
    }
});