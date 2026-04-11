import { z } from "zod";

// Example schema — replace with your domain types.

export const insertItemSchema = z.object({
  title: z.string().default(""),
  body: z.string().default(""),
});

export const updateItemSchema = insertItemSchema.partial();

export type InsertItem = z.infer<typeof insertItemSchema>;
export type UpdateItem = z.infer<typeof updateItemSchema>;

export interface Item {
  id: string;
  title: string;
  body: string;
  createdAt: string;
  updatedAt: string;
}
