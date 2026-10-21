// Checks the Feature-Sliced Design rules on src/:
//   - a layer imports only from the layers below it
//     (app → pages → widgets → features → entities → shared);
//   - slices of one layer don't import each other;
//   - other slices are reached through their public index, never a deep path.
// Exits 1 on any violation.
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative, sep } from "node:path";

const root = new URL("../src", import.meta.url).pathname;
const layers = ["app", "pages", "widgets", "features", "entities", "shared"];
// app and shared have segments, not slices.
const sliced = new Set(["pages", "widgets", "features", "entities"]);

function files(dir) {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name);
    if (statSync(path).isDirectory()) return files(path);
    return /\.tsx?$/.test(name) ? [path] : [];
  });
}

const problems = [];
for (const file of files(root)) {
  const [layer, slice] = relative(root, file).split(sep);
  const source = readFileSync(file, "utf8");
  for (const [, spec] of source.matchAll(/(?:from|import)\s*\(?\s*["']@\/([^"']+)["']/g)) {
    const [toLayer, toSlice, ...rest] = spec.split("/");
    const where = `${relative(root, file)} → @/${spec}`;
    if (!layers.includes(toLayer)) {
      problems.push(`${where}: unknown layer`);
    } else if (layers.indexOf(toLayer) < layers.indexOf(layer)) {
      problems.push(`${where}: ${layer} can't import ${toLayer}`);
    } else if (toLayer === layer && sliced.has(layer) && toSlice !== slice) {
      problems.push(`${where}: slices of ${layer} can't import each other`);
    } else if (toLayer === layer && sliced.has(layer)) {
      problems.push(`${where}: import your own slice by a relative path`);
    } else if (sliced.has(toLayer) && rest.length > 0) {
      problems.push(`${where}: import the slice's public index`);
    }
  }
}
if (problems.length) {
  console.log(problems.join("\n"));
  process.exit(1);
}
console.log("FSD: no violations");
