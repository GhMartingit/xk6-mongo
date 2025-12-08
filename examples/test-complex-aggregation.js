import xk6_mongo from 'k6/x/mongo';

const client = xk6_mongo.newClient('mongodb://localhost:27017');

export default () => {
  // Complex aggregation pipeline with multiple stages
  // This example demonstrates:
  // - Filtering documents
  // - Grouping and aggregation
  // - Sorting results
  // - Limiting output
  // - Projecting specific fields

  const pipeline = [
    // Stage 1: Match documents from the last 30 days
    {
      $match: {
        createdAt: {
          $gte: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000)
        },
        status: "active"
      }
    },

    // Stage 2: Group by category and calculate statistics
    {
      $group: {
        _id: "$category",
        totalCount: { $sum: 1 },
        avgPrice: { $avg: "$price" },
        maxPrice: { $max: "$price" },
        minPrice: { $min: "$price" },
        totalRevenue: { $sum: "$revenue" },
        products: { $push: "$name" }
      }
    },

    // Stage 3: Add computed fields
    {
      $addFields: {
        category: "$_id",
        priceRange: { $subtract: ["$maxPrice", "$minPrice"] },
        avgRevenuePerItem: { $divide: ["$totalRevenue", "$totalCount"] }
      }
    },

    // Stage 4: Filter groups with significant revenue
    {
      $match: {
        totalRevenue: { $gt: 1000 }
      }
    },

    // Stage 5: Sort by total revenue (descending)
    {
      $sort: { totalRevenue: -1 }
    },

    // Stage 6: Limit to top 10 categories
    {
      $limit: 10
    },

    // Stage 7: Project final output fields
    {
      $project: {
        _id: 0,
        category: 1,
        totalCount: 1,
        avgPrice: { $round: ["$avgPrice", 2] },
        totalRevenue: { $round: ["$totalRevenue", 2] },
        avgRevenuePerItem: { $round: ["$avgRevenuePerItem", 2] },
        priceRange: { $round: ["$priceRange", 2] },
        topProducts: { $slice: ["$products", 5] }  // Show only top 5 products
      }
    }
  ];

  const results = client.aggregate("testdb", "products", pipeline);

  console.log(`Found ${results.length} top performing categories:`);
  results.forEach((category, index) => {
    console.log(`${index + 1}. ${category.category}:`);
    console.log(`   Items: ${category.totalCount}, Revenue: $${category.totalRevenue}`);
    console.log(`   Avg Price: $${category.avgPrice}, Avg Revenue/Item: $${category.avgRevenuePerItem}`);
  });
}

// Example with $lookup (join) for relational data
export function complexJoinExample() {
  const pipeline = [
    // Match orders from last week
    {
      $match: {
        orderDate: {
          $gte: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)
        }
      }
    },

    // Join with customers collection
    {
      $lookup: {
        from: "customers",
        localField: "customerId",
        foreignField: "_id",
        as: "customerInfo"
      }
    },

    // Unwind the customer array
    {
      $unwind: "$customerInfo"
    },

    // Join with products collection
    {
      $lookup: {
        from: "products",
        localField: "productId",
        foreignField: "_id",
        as: "productInfo"
      }
    },

    // Unwind products
    {
      $unwind: "$productInfo"
    },

    // Group by customer and calculate totals
    {
      $group: {
        _id: "$customerId",
        customerName: { $first: "$customerInfo.name" },
        customerEmail: { $first: "$customerInfo.email" },
        totalOrders: { $sum: 1 },
        totalSpent: { $sum: "$productInfo.price" },
        products: { $push: "$productInfo.name" }
      }
    },

    // Sort by total spent
    {
      $sort: { totalSpent: -1 }
    },

    // Format output
    {
      $project: {
        _id: 0,
        customerName: 1,
        customerEmail: 1,
        totalOrders: 1,
        totalSpent: { $round: ["$totalSpent", 2] },
        avgOrderValue: {
          $round: [{ $divide: ["$totalSpent", "$totalOrders"] }, 2]
        },
        uniqueProducts: { $size: "$products" }
      }
    }
  ];

  const results = client.aggregate("testdb", "orders", pipeline);
  console.log(`Top customers this week: ${JSON.stringify(results, null, 2)}`);
}
