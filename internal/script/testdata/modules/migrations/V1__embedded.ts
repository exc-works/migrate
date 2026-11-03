import { value } from "../shared/value.ts";
export function up() {
  if (value !== 42) throw Error("embedded helper was not loaded");
}
