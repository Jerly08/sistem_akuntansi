import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	env: {
		NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
	},
	typescript: {
		ignoreBuildErrors: true,
	},
	eslint: {
		ignoreDuringBuilds: true,
	},
	// Konfigurasi Webpack hanya untuk mode non-Turbopack
	...(!process.env.TURBOPACK && {
		webpack: (config, { isServer }) => {
			// Handle client-side module resolution for jsPDF
			if (!isServer) {
				config.resolve.fallback = {
					...config.resolve.fallback,
					fs: false,
					path: false,
				};
				
				// Handle jsPDF ES modules
				config.module.rules.push({
					test: /\.m?js$/,
					resolve: {
						fullySpecified: false
					}
				});
			}
			return config;
		},
	}),
	async rewrites() {
		return [
			{
				source: '/api/:path*',
				destination: 'http://localhost:8080/api/:path*',
			},
		];
	},
};

export default nextConfig;
