const ID_PREFIX = "org-state";

const randomSegment = () => Math.random().toString(36).slice(2, 8);

export const generateIdempotencyKey = (
  action: "suspend" | "activate",
  code: string,
): string => {
  const normalizedCode = code.trim();
  const timestamp = new Date()
    .toISOString()
    .replace(/[-:T.Z]/g, "")
    .slice(0, 14);
  return `${ID_PREFIX}-${action}-${normalizedCode}-${timestamp}-${randomSegment()}`;
};

export default generateIdempotencyKey;
