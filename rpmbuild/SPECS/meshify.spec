Name:           meshify
Version:        1.11
Release:        1%{?dist}
Summary:        Meshify Client for RPM based linux systems

License:        WTF
URL:            https://meshify.app
BuildRoot:      ~/rpmbuild/

%description
Meshify Agent for RPM based linux distributions

%prep
################################################################################
# Create the build tree and copy the files from the development directories    #
# into the build tree.                                                         #
################################################################################
echo "BUILDROOT = $RPM_BUILD_ROOT"
mkdir -p $RPM_BUILD_ROOT/usr/bin/
mkdir -p $RPM_BUILD_ROOT/etc/meshify/
mkdir -p $RPM_BUILD_ROOT/lib/systemd/system/

cp /home/opc/go/src/meshify-client/meshify-client $RPM_BUILD_ROOT/usr/bin
cp /home/opc/go/src/meshify-client/rpmbuild/BUILD/lib/systemd/system/meshify.service $RPM_BUILD_ROOT/lib/systemd/system/meshify.service
exit


%files
%attr(0744, root, root) /usr/bin/meshify-client
%attr(0644, root, root) /lib/systemd/system/meshify.service
%doc


%clean
rm -rf $RPM_BUILD_ROOT/usr/
rm -rf $RPM_BUILD_ROOT/lib/
rm -rf $RPM_BUILD_ROOT/etc/

%post
/usr/bin/systemctl enable meshify.service > /dev/null 2>&1
/usr/bin/systemctl start meshify.service
exit 0
%preun
/usr/bin/systemctl stop meshify.service
exit 0


%changelog
* Tue Nov 16 2021 by ALan Graham
Support client managed private keys
* Tue Nov 09 2021 by Alan Graham
Improve wireguard conf file generation
* Thu Aug 26 2021 by Alan Graham
Initial Release
