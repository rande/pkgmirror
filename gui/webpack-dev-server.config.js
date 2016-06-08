const webpack               = require('webpack');
const path                  = require('path');
const buildPath             = path.resolve(__dirname, 'build');
const nodeModulesPath       = path.resolve(__dirname, 'node_modules');
const TransferWebpackPlugin = require('transfer-webpack-plugin');
const WriteFilePlugin       = require('write-file-webpack-plugin');

console.log(buildPath);

const config = {
    // Entry points to the project
    entry:     [
        'webpack/hot/dev-server',
        'webpack/hot/only-dev-server',
        path.join(__dirname, '/src/app.js'),
    ],
    // Server Configuration options
    devServer: {
        contentBase: 'src', // Relative directory for base of server
        devtool:     'eval',
        hot:         true, // Live-reload
        inline:      true,
        port:        3000, // Port Number
        host:        'localhost', // Change to '0.0.0.0' for external facing server
        proxy:       {
            '*': {
                target: 'http://localhost:8000',
                secure: false
            }
        },
        outputPath: buildPath,
    },
    devtool:   'eval',
    output:    {
        path:     buildPath, // Path of output file
        filename: 'app.js',
    },
    plugins:   [
        // Enables Hot Modules Replacement
        new webpack.HotModuleReplacementPlugin(),
        // Allows error warnings but does not stop compiling.
        new webpack.NoErrorsPlugin(),
        // Moves files
        new TransferWebpackPlugin([{from: 'static'},], path.resolve(__dirname, 'src')),
        // Force write to disk, so the gobindata can catchup files on run
        new WriteFilePlugin()
    ],
    module:    {
        loaders: [
            {
                // React-hot loader and
                test:    /\.js$/, // All .js files
                loaders: ['react-hot', 'babel-loader'], // react-hot is like browser sync and babel loads jsx and es6-7
                exclude: [nodeModulesPath],
            },
        ],
    }
};

module.exports = config;
