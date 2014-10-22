
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
	config.vm.box = "mistify/trusty64-vmware"
	config.vm.box_url  = "http://www.akins.org/boxes/mistify-ubuntu-vmware.box"
	config.ssh.forward_agent = true

	config.vm.synced_folder ".", "/home/vagrant/go/src/github.com/mistifyio/mistify-agent", create: true

    config.vm.network "forwarded_port", guest: 8080, host: 8080, auto_correct: true

	config.vm.provider "vmware_fusion" do |v|
		# GUI is needed because when you use bridging inside Linux,
		# Fusion must ask for admin password
		v.gui = true
		v.vmx["memsize"] = "1024"
		v.vmx["numvcpus"] = "2"
		v.vmx["vhv.enable"] = "TRUE"
	end

    config.vm.provision "shell", privileged: false, inline: <<EOF
cd /home/vagrant/go/src/github.com/mistifyio/mistify-agent
go get github.com/tools/godep
godep go install
EOF
end
