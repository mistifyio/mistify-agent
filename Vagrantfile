
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
	config.vm.box = "ubuntu/trusty64"
	config.ssh.forward_agent = true
	config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/mistifyio/mistify-agent", create: true
    config.vm.network "forwarded_port", guest: 19999, host: 19999, auto_correct: true

    config.vm.provision "shell", privileged: true, inline: <<EOS
if [ ! -x /usr/local/go/bin/go ]; then
     apt-get update
     apt-get -y install curl git-core mercurial
     cd /tmp
     curl -s -L -O http://golang.org/dl/go1.3.3.linux-amd64.tar.gz
     tar -C /usr/local -zxf go1.3.3.linux-amd64.tar.gz
     rm /tmp/go1.3.3.linux-amd64.tar.gz
     chown -R vagrant /home/vagrant/go
fi

if [ ! -x /usr/local/bin/jq ]; then
    cd /tmp
    curl -s -L -O http://stedolan.github.io/jq/download/linux64/jq
    mv jq /usr/local/bin
    chmod 0555 /usr/local/bin/jq
fi

cat <<EOF > /etc/profile.d/go.sh
GOPATH=\$HOME/go
export GOPATH
PATH=\$GOPATH/bin:\$PATH:/usr/local/go/bin
export PATH
EOF

EOS
end
