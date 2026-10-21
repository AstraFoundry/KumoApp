import { isCallError } from "mygo-runtime";

/** The message of an error from Go or the page. */
export function errorMessage(err: unknown): string {
  if (isCallError(err)) return err.message;
  if (err instanceof Error) return err.message;
  return String(err);
}
