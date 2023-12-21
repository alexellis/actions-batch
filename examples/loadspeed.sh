#!/bin/bash

mkdir -p output

# https://github.com/actions/runner-images/blob/main/images/ubuntu/scripts/build/install-phantomjs.sh
# Install required dependencies
sudo apt-get install -yq chrpath libssl-dev libxft-dev libfreetype6 libfreetype6-dev libfontconfig1 libfontconfig1-dev

# Define the version and hash of PhantomJS to be installed
dir_name=phantomjs-2.1.1-linux-x86_64
download_url="https://bitbucket.org/ariya/phantomjs/downloads/$dir_name.tar.bz2"
archive_path=/tmp/$dir_name.tar.bz2

curl -Lsfo "$archive_path" "$download_url"

# Extract the archive and create a symbolic link to the executable
sudo tar xjf "$archive_path" -C /usr/local/share
ln -sf /usr/local/share/$dir_name/bin/phantomjs /usr/local/bin
sudo chmod +x /usr/local/bin/phantomjs

cat >> loadspeed.js << EOF
var page = require('webpage').create(),
  system = require('system'),
  t, address;

if (system.args.length === 1) {
  console.log('Usage: loadspeed.js [some URL]');
  phantom.exit();
}

page.onError = function(msg, trace) {
  var msgStack = ['ERROR: ' + msg];

  if (trace && trace.length) {
    msgStack.push('TRACE:');
    trace.forEach(function(t) {
      msgStack.push(' -> ' + t.file + ': ' + t.line + (t.function ? ' (in function "' + t.function +'")' : ''));
    });
  }

  console.error(msgStack.join('\n'));
};

page.settings.userAgent = 'Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2062.120 Safari/537.36';
page.javascriptEnabled = true;

t = Date.now();
address = system.args[1];
page.open(address, function(status) {
  if (status !== 'success') {
    console.log('FAIL to load the address: ' + address);
  } else {
    t = Date.now() - t;
    console.log('Loading ' + system.args[1]);
    console.log('Loading time ' + t + ' msec');

    // Save a screenshot to output folder
    page.render('output/page.png');
  }
  phantom.exit();
});
EOF

cat loadspeed.js

phantomjs loadspeed.js "http://www.google.com/"
