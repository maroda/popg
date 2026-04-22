const sectors = [
    {color:"#ca2652", label:"Adaptive Capacity"},
    {color:"#79751b", label:"FRAM"},
    {color:"#a23737", label:"ETTO"},
    {color:"#2144f3", label:"Robustness"},
    {color:"#683572", label:"RAG"},
    {color:"#0a7d4a", label:"Human Work"},
    {color:"#0e4467", label:"Incident Review"},
];

const rand = (m, M) => Math.random() * (M - m) + m;
const tot = sectors.length;
const EL_spin = document.querySelector("#spin");
const ctx = document.querySelector("#wheel").getContext('2d');
const dia = ctx.canvas.width;
const rad = dia / 2;
const PI = Math.PI;
const TAU = 2 * PI;
const arc = TAU / sectors.length;

const friction = 0.991; // 0.995=soft, 0.99=mid, 0.98=hard
let angVel = 0; // Angular velocity
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

function rotate() {
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
    if (angVel < 0.002) angVel = 0; // Bring to stop
    ang += angVel; // Update angle
    ang %= TAU; // Normalize angle
    rotate();
}

function engine() {
    frame();
    requestAnimationFrame(engine)
}

// INIT
let randseed = 0.35 // <<< spindata.RandSeed
sectors.forEach(drawSector);
rotate(); // Initial rotation
engine(); // Start engine
EL_spin.addEventListener("click", () => {
    if (!angVel) angVel = rand(0.25, randseed); // <<< spindata.Velocity
});
