const { decode } = require("fast-png");
const jsQR = require("jsqr");

function toRGBA(png) {
  const { data, width, height, channels, depth } = png;
  const rgba = new Uint8ClampedArray(width * height * 4);

  if (depth === 1) {
    const rowBytes = Math.ceil(width / 8);
    for (let y = 0; y < height; y++) {
      for (let x = 0; x < width; x++) {
        const val =
          (data[y * rowBytes + (x >> 3)] >> (7 - (x & 7))) & 1 ? 255 : 0;
        const i = (y * width + x) * 4;
        rgba[i] = val;
        rgba[i + 1] = val;
        rgba[i + 2] = val;
        rgba[i + 3] = 255;
      }
    }
    return rgba;
  }

  if (channels === 4) return new Uint8ClampedArray(data);
  for (let i = 0; i < width * height; i++) {
    const r = channels === 1 ? data[i] : data[i * channels];
    const g = channels === 1 ? data[i] : data[i * channels + 1];
    const b = channels === 1 ? data[i] : data[i * channels + 2];
    rgba[i * 4] = r;
    rgba[i * 4 + 1] = g;
    rgba[i * 4 + 2] = b;
    rgba[i * 4 + 3] = 255;
  }
  return rgba;
}

function decodeQR(buf) {
  const png = decode(buf, { checkCrc: false });
  const qr = jsQR(toRGBA(png), png.width, png.height);
  return qr ? qr.data : null;
}

module.exports = { decodeQR };
