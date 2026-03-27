# Catalogizer -- Collection User Guide

## Table of Contents

1. [What Are Collections](#what-are-collections)
2. [Creating a Collection](#creating-a-collection)
3. [Smart Collections with Rules](#smart-collections-with-rules)
4. [Collection Templates](#collection-templates)
5. [Adding and Removing Items](#adding-and-removing-items)
6. [Sharing Collections](#sharing-collections)
7. [Collection Analytics](#collection-analytics)
8. [Bulk Operations](#bulk-operations)
9. [Export and Import](#export-and-import)
10. [Troubleshooting](#troubleshooting)

---

## What Are Collections

Collections are user-defined groupings of media entities that let you organize your library beyond the automatic type-based categories. While Catalogizer automatically categorizes your media by type (movies, TV shows, music, etc.) and builds hierarchies (show > season > episode), collections give you complete freedom to group entities however you want.

Examples of collections:

- "Saturday Movie Night" -- A curated list of movies for family viewing.
- "90s Nostalgia" -- Movies, TV shows, and music from the 1990s.
- "Study Playlist" -- A mix of ambient music albums and lo-fi tracks.
- "Award Winners" -- Oscar-winning films collected over the years.
- "Kids Approved" -- Age-appropriate content for children.

Collections are stored in the `media_collections` and `media_collection_items` database tables. Each collection belongs to the user who created it, though collections can be shared with other users.

---

## Creating a Collection

### From the Web Interface

1. Navigate to **Collections** in the main navigation sidebar.
2. Click the **New Collection** button in the top toolbar.
3. Fill in the collection details:
   - **Name** -- A descriptive name for the collection (required).
   - **Description** -- An optional summary of the collection's purpose or theme.
   - **Cover image** -- Choose from your library posters or upload a custom image.
   - **Visibility** -- Private (only you) or Shared (visible to other users).
4. Click **Create**.

The new collection appears in your collection list, initially empty. You can start adding items immediately.

### From the Entity Detail Page

1. Open any entity detail page.
2. Click the three-dot menu or the **Add to Collection** button.
3. Select an existing collection from the dropdown, or click **New Collection** to create one on the spot.
4. The entity is added to the selected collection immediately.

### From the Entity Browser

1. Right-click an entity card in the Browse view.
2. Select **Add to Collection** from the context menu.
3. Choose the target collection or create a new one.

On the desktop app, you can also drag and drop entity cards directly onto a collection in the sidebar.

---

## Smart Collections with Rules

Smart collections are dynamic collections that automatically include entities matching a set of rules. Instead of manually adding items, you define criteria, and Catalogizer keeps the collection up to date as your library changes.

### Creating a Smart Collection

1. Navigate to **Collections** and click **New Collection**.
2. Toggle the **Smart Collection** switch to enable rule-based filtering.
3. Define one or more rules using the rule builder.

### Rule Builder

Each rule consists of a **field**, an **operator**, and a **value**:

| Field | Operators | Example |
|-------|-----------|---------|
| **Type** | is, is not | Type is "movie" |
| **Year** | is, is not, greater than, less than, between | Year between 1990 and 1999 |
| **Rating** | greater than, less than, equal to | Rating greater than 7.5 |
| **Genre** | contains, does not contain | Genre contains "Sci-Fi" |
| **Title** | contains, starts with, ends with | Title contains "Star" |
| **Date added** | before, after, in the last | Date added in the last 30 days |
| **File count** | greater than, less than, equal to | File count greater than 1 |
| **Has metadata** | is true, is false | Has metadata is true |

### Combining Rules

Rules are combined using **AND** (all rules must match) or **OR** (any rule must match) logic. Click the logic toggle between rules to switch. For complex conditions, you can create nested groups of rules.

Example -- "Recent High-Rated Sci-Fi Movies":
- Type **is** "movie" **AND**
- Genre **contains** "Sci-Fi" **AND**
- Rating **greater than** 7.0 **AND**
- Year **greater than** 2020

### Auto-Update Behavior

Smart collections refresh automatically when:

- A new scan completes and new entities are aggregated.
- An entity's metadata is updated (e.g., rating changes after a provider refresh).
- You manually refresh the collection by clicking the refresh icon.

Smart collection membership is read-only -- you cannot manually add or remove items. To convert a smart collection to a manual collection, click **Convert to Manual** in the collection settings. This snapshots the current membership and removes the rules.

---

## Collection Templates

Collection templates provide pre-built configurations for common collection patterns. They save time by pre-filling the collection name, description, and smart rules.

### Available Templates

| Template | Description | Rules |
|----------|-------------|-------|
| **Recently Added** | Items added in the last 7 days | Date added in the last 7 days |
| **Top Rated Movies** | Highest-rated films in your library | Type is movie, Rating > 8.0 |
| **Unwatched** | Items you have not played yet | Play count is 0 |
| **Duplicates** | Entities with more than one linked file | File count > 1 |
| **Missing Metadata** | Items without external provider data | Has metadata is false |
| **4K Collection** | Files available in 4K resolution | Resolution is 2160p |
| **Music Discovery** | Albums added in the last 30 days | Type is music_album, Date added in last 30 days |

### Using a Template

1. Navigate to **Collections** and click **New Collection**.
2. Click the **Use Template** dropdown.
3. Select a template. The form is pre-filled with the template's name, description, and rules.
4. Modify any fields as needed (you are not locked into the template's defaults).
5. Click **Create**.

---

## Adding and Removing Items

### Adding Items to a Manual Collection

There are multiple ways to add entities to a collection:

- **From the entity detail page**: Click **Add to Collection** and select the target collection.
- **From the entity browser**: Right-click a card and select **Add to Collection**, or use the multi-select mode (see Bulk Operations below) to add multiple items at once.
- **From the collection page**: Click **Add Items**, which opens a search dialog. Search for entities by title and click the plus icon to add them.
- **Drag and drop** (desktop app): Drag entity cards from the browser onto a collection name in the sidebar.

### Reordering Items

Items in a manual collection can be reordered:

1. Open the collection detail page.
2. Click the **Reorder** button in the toolbar.
3. Drag items up or down to the desired position.
4. Click **Save Order** when finished.

The order is preserved and displayed consistently across all platforms (web, desktop, mobile, TV).

### Removing Items

- **From the collection page**: Hover over an item card and click the remove (X) icon.
- **From the entity detail page**: In the collections section, click the remove icon next to the collection name.
- **Bulk removal**: Use multi-select mode on the collection page, select the items to remove, and click **Remove Selected**.

Removing an item from a collection does not delete the entity or its files. It only removes the association between the entity and the collection.

---

## Sharing Collections

Collections can be shared with other Catalogizer users on the same server.

### Visibility Settings

| Visibility | Who Can See | Who Can Edit |
|------------|-------------|-------------|
| **Private** | Only the creator | Only the creator |
| **Shared (read-only)** | All users on the server | Only the creator |
| **Shared (collaborative)** | All users on the server | All users on the server |

### Changing Visibility

1. Open the collection detail page.
2. Click the gear icon to open collection settings.
3. Under **Visibility**, select the desired option.
4. Click **Save**.

### Shared Collection Behavior

When a collection is shared, other users see it in their **Collections** sidebar under a "Shared with me" section. Collaborative collections allow any user to add or remove items, but only the creator can delete the collection itself or change its settings.

---

## Collection Analytics

The collection detail page includes an analytics section that provides insights about the collection's contents.

### Statistics

- **Total items** -- Number of entities in the collection.
- **Total files** -- Number of physical files linked to all entities in the collection.
- **Total size** -- Combined file size of all linked files.
- **Type breakdown** -- A chart showing the distribution of media types in the collection (e.g., 60% movies, 25% TV episodes, 15% music).

### Rating Distribution

A histogram showing the distribution of ratings across items in the collection. Useful for identifying whether the collection skews toward high-rated or low-rated content.

### Year Distribution

A timeline chart showing how many items in the collection were released in each year or decade. Useful for understanding the temporal spread of the collection.

### Metadata Coverage

A progress bar showing what percentage of items in the collection have external metadata from providers. Items without metadata may benefit from a manual metadata refresh.

---

## Bulk Operations

Bulk operations allow you to perform actions on multiple entities or collections at once.

### Multi-Select Mode

1. In the entity browser or on a collection page, click the **Select** button in the toolbar (or press and hold an entity card on mobile).
2. Click entity cards to select or deselect them. Selected cards display a checkmark overlay.
3. The toolbar shows the count of selected items and available bulk actions.

### Available Bulk Actions

| Action | Description |
|--------|-------------|
| **Add to Collection** | Add all selected entities to a collection |
| **Remove from Collection** | Remove all selected entities from the current collection |
| **Move to Collection** | Remove from the current collection and add to another |
| **Set Rating** | Apply the same user rating to all selected entities |
| **Add Tag** | Apply a tag to all selected entities |
| **Refresh Metadata** | Trigger a metadata refresh for all selected entities |

### Bulk Collection Management

From the Collections list page, you can also perform bulk operations on collections themselves:

- **Delete multiple collections** -- Select collections and click **Delete Selected**.
- **Export multiple collections** -- Select collections and click **Export Selected** to download them as a single file.

---

## Export and Import

Collections can be exported and imported for backup, migration, or sharing with users on different Catalogizer servers.

### Export Format

Collections are exported as JSON files containing:

- Collection metadata (name, description, visibility, rules for smart collections).
- A list of member entities identified by title, type, and year (not by internal database IDs, which differ between servers).

### Exporting a Collection

1. Open the collection detail page.
2. Click the gear icon and select **Export**.
3. Choose the export format: **JSON** (machine-readable) or **Markdown** (human-readable list).
4. The file downloads to your browser's default download location.

### Importing a Collection

1. Navigate to **Collections** and click **Import Collection**.
2. Select the JSON file to import.
3. The import dialog shows a preview of the collection: name, description, and the list of entities it references.
4. Catalogizer attempts to match each referenced entity to an existing entity in your library by title, type, and year.
5. Matched entities are displayed with a green checkmark. Unmatched entities (items in the export that do not exist in your library) are displayed with a warning icon.
6. Click **Import** to create the collection with all matched entities. Unmatched references are skipped.

### API Access

Collections can also be managed programmatically via the REST API:

```
GET    /api/v1/collections                  -- List all collections
POST   /api/v1/collections                  -- Create a collection
GET    /api/v1/collections/:id              -- Get collection details and items
PUT    /api/v1/collections/:id              -- Update collection metadata
DELETE /api/v1/collections/:id              -- Delete a collection
POST   /api/v1/collections/:id/items        -- Add items to a collection
DELETE /api/v1/collections/:id/items/:itemId -- Remove an item from a collection
```

---

## Troubleshooting

### Collection Not Updating (Smart Collections)

- Smart collections refresh when scans complete or metadata changes. If the collection seems stale, click the refresh icon on the collection page to force a re-evaluation of the rules.
- Verify that the rules are correct. Open collection settings and review each rule's field, operator, and value.
- Check that the entities you expect to appear actually match all the rules (if using AND logic) or at least one rule (if using OR logic).

### Items Missing After Import

- Imported collections match entities by title, type, and year. If your library uses slightly different naming (e.g., "The Matrix" vs. "Matrix, The"), the match may fail.
- Re-scan your library to ensure all entities are up to date, then try the import again.

### Cannot Share a Collection

- Only the collection creator can change visibility settings. If you did not create the collection, you cannot share it.
- Verify that other users exist on the server. Sharing is only meaningful when there are multiple user accounts.

### Collection Order Not Preserved

- Collection item order is only supported for manual collections. Smart collections always display items according to their rule-based sort order.
- Ensure you clicked **Save Order** after reordering. Navigating away without saving discards changes.

### Bulk Operation Fails Partway

- If a bulk operation (such as adding 100 items to a collection) fails after processing some items, the successfully processed items are retained. Check the error message for the cause of the failure (typically a network timeout or server resource limit) and retry the operation for the remaining items.
